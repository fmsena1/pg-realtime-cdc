package http

import (
	"net/http"

	"github.com/fmsena1/pg-realtime-cdc/backend/ws"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func ServeWS(hub *ws.Hub, w http.ResponseWriter, r *http.Request) {
	conn, _ := upgrader.Upgrade(w, r, nil)
	ch := make(chan []byte)

	hub.Clients[ch] = true

	go func() {
		for msg := range ch {
			conn.WriteMessage(websocket.TextMessage, msg)
		}
	}()
}
