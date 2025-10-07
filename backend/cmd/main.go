package main

import (
	"log"
	"net/http"

	wb "backend/internal/api/websocket"
	"backend/internal/application"
	rabbitmq "backend/internal/infrastructure"
)

func main() {
	publisher, err := rabbitmq.GetRPCconnection("amqp://guest:guest@localhost:5673/")
	if err != nil {
		log.Fatalf("Failed to initialize rabbitmq connection: %v", err)
	}

	app := application.NewApp(publisher)
	wsHandler := wb.NewWSHandler(app)

	http.HandleFunc("/ws", wsHandler.WsHandler)
	http.Handle("/", http.FileServer(http.Dir("./cmd/public")))

	log.Println("Server starting on :8080")
	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatalf("ListenAndServe error: %v", err)
	}
}
