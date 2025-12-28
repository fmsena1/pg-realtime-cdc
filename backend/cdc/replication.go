package cdc

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/jackc/pglogrepl"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgproto3"
)

type Event struct {
	Table string                 `json:"table"`
	Op    string                 `json:"op"`
	Data  map[string]interface{} `json:"data"`
	Old   map[string]interface{} `json:"old,omitempty"`
}

var relations = make(map[uint32]*pglogrepl.RelationMessage)

func Start(hub chan []byte) {
	ctx := context.Background()

	connConfig, err := pgconn.ParseConfig(
		"postgres://realtime:realtime@postgres:5432/realtime?replication=database",
	)
	if err != nil {
		log.Fatal("parse config:", err)
	}

	conn, err := pgconn.ConnectConfig(ctx, connConfig)
	if err != nil {
		log.Fatal("connect:", err)
	}
	defer conn.Close(ctx)

	slot := "go_realtime_slot"

	err = pglogrepl.StartReplication(
		ctx,
		conn,
		slot,
		0,
		pglogrepl.StartReplicationOptions{
			PluginArgs: []string{
				"proto_version '1'",
				"publication_names 'go_realtime_pub'",
			},
		},
	)
	if err != nil {
		log.Fatal("start replication:", err)
	}

	log.Println("📡 Logical replication started")

	var lastLSN pglogrepl.LSN
	standbyTimeout := 10 * time.Second
	nextStatus := time.Now().Add(standbyTimeout)

	for {
		ctxTimeout, cancel := context.WithDeadline(ctx, nextStatus)
		msg, err := conn.ReceiveMessage(ctxTimeout)
		cancel()

		if err != nil {
			if pgconn.Timeout(err) {
				err = pglogrepl.SendStandbyStatusUpdate(ctx, conn, pglogrepl.StandbyStatusUpdate{
					WALWritePosition: lastLSN,
					WALFlushPosition: lastLSN,
					WALApplyPosition: lastLSN,
					ClientTime:       time.Now(),
				})
				if err != nil {
					log.Println("standby timeout ack error:", err)
					return
				}
				nextStatus = time.Now().Add(standbyTimeout)
				continue
			}

			log.Println("receive error:", err)
			return
		}

		switch m := msg.(type) {
		case *pgproto3.CopyData:
			switch m.Data[0] {
			case pglogrepl.XLogDataByteID:
				xlog, err := pglogrepl.ParseXLogData(m.Data[1:])
				if err != nil {
					log.Println("parse xlog error:", err)
					continue
				}

				lastLSN = xlog.WALStart + pglogrepl.LSN(len(xlog.WALData))

				msg, err := pglogrepl.Parse(xlog.WALData)
				if err != nil {
					log.Printf("parse message error: %v", err)
					continue
				}

				var event *Event
				switch v := msg.(type) {
				case *pglogrepl.RelationMessage:
					relations[v.RelationID] = v
					log.Printf("Relation: %s.%s (ID: %d)", v.Namespace, v.RelationName, v.RelationID)
					continue

				case *pglogrepl.InsertMessage:
					event = parseInsert(v)

				case *pglogrepl.UpdateMessage:
					event = parseUpdate(v)

				case *pglogrepl.DeleteMessage:
					event = parseDelete(v)
				}

				if event != nil {
					payload, _ := json.Marshal(event)
					hub <- payload
				}

				err = pglogrepl.SendStandbyStatusUpdate(ctx, conn, pglogrepl.StandbyStatusUpdate{
					WALWritePosition: lastLSN,
					WALFlushPosition: lastLSN,
					WALApplyPosition: lastLSN,
					ClientTime:       time.Now(),
				})
				if err != nil {
					log.Println("standby ack error:", err)
					return
				}

				nextStatus = time.Now().Add(standbyTimeout)

			case pglogrepl.PrimaryKeepaliveMessageByteID:
				keepalive, err := pglogrepl.ParsePrimaryKeepaliveMessage(m.Data[1:])
				if err != nil {
					log.Println("parse keepalive error:", err)
					continue
				}

				if keepalive.ReplyRequested {
					err = pglogrepl.SendStandbyStatusUpdate(ctx, conn, pglogrepl.StandbyStatusUpdate{
						WALWritePosition: lastLSN,
						WALFlushPosition: lastLSN,
						WALApplyPosition: lastLSN,
						ClientTime:       time.Now(),
					})
					if err != nil {
						log.Println("keepalive ack error:", err)
						return
					}
				}
			}
		}
	}
}

func parseInsert(msg *pglogrepl.InsertMessage) *Event {
	rel, ok := relations[msg.RelationID]
	if !ok {
		log.Printf("Unknown relation ID: %d", msg.RelationID)
		return nil
	}

	event := &Event{
		Table: rel.RelationName,
		Op:    "INSERT",
		Data:  make(map[string]interface{}),
	}

	parseTuple(msg.Tuple, rel, event.Data)
	return event
}

func parseUpdate(msg *pglogrepl.UpdateMessage) *Event {
	rel, ok := relations[msg.RelationID]
	if !ok {
		log.Printf("Unknown relation ID: %d", msg.RelationID)
		return nil
	}

	event := &Event{
		Table: rel.RelationName,
		Op:    "UPDATE",
		Data:  make(map[string]interface{}),
		Old:   make(map[string]interface{}),
	}

	if msg.OldTuple != nil {
		parseTuple(msg.OldTuple, rel, event.Old)
	}

	if msg.NewTuple != nil {
		parseTuple(msg.NewTuple, rel, event.Data)
	}

	return event
}

func parseDelete(msg *pglogrepl.DeleteMessage) *Event {
	rel, ok := relations[msg.RelationID]
	if !ok {
		log.Printf("Unknown relation ID: %d", msg.RelationID)
		return nil
	}

	event := &Event{
		Table: rel.RelationName,
		Op:    "DELETE",
		Data:  make(map[string]interface{}),
	}

	if msg.OldTuple != nil {
		parseTuple(msg.OldTuple, rel, event.Data)
	}

	return event
}

func parseTuple(tuple *pglogrepl.TupleData, rel *pglogrepl.RelationMessage, result map[string]interface{}) {
	for i, col := range tuple.Columns {
		if i >= len(rel.Columns) {
			continue
		}

		colName := rel.Columns[i].Name
		if col.DataType == 'n' {
			result[colName] = nil
		} else if col.DataType == 'u' {
			result[colName] = "[unchanged TOAST]"
		} else {

			result[colName] = string(col.Data)
		}
	}
}
