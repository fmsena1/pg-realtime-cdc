package main

import (
	"log"
	"net/http"

	"github.com/fmsena1/pg-realtime-cdc/backend/api"
	"github.com/fmsena1/pg-realtime-cdc/backend/cdc"
	"github.com/fmsena1/pg-realtime-cdc/backend/db"
	httpSrv "github.com/fmsena1/pg-realtime-cdc/backend/http"
	"github.com/fmsena1/pg-realtime-cdc/backend/ws"
)

func main() {
	if err := db.Init(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	hub := ws.NewHub()
	go hub.Run()
	go cdc.Start(hub.Broadcast)

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		httpSrv.ServeWS(hub, w, r)
	})

	http.HandleFunc("/api/messages", api.CORS(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			api.GetMessages(w, r)
		case http.MethodPost:
			api.CreateMessage(w, r)
		case http.MethodOptions:

			return
		default:
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))

	http.HandleFunc("/api/messages/update", api.CORS(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPut:
			api.UpdateMessage(w, r)
		case http.MethodOptions:
			return
		default:
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))

	http.HandleFunc("/api/messages/delete", api.CORS(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodDelete:
			api.DeleteMessage(w, r)
		case http.MethodOptions:
			return
		default:
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))

	log.Println("🚀 Go Realtime on :8080")
	log.Println("📡 WebSocket: ws://localhost:8080/ws")
	log.Println("📡 REST API: http://localhost:8080/api/messages")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
