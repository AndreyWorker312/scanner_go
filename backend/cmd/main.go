package main

import (
	wb "backend/internal/api/websocket"
	"backend/internal/application"
	rabbitmq "backend/internal/infrastructure"
	"log"
	"net/http"
)

func main() {
	publisher, err := rabbitmq.GetRPCconnection("amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Fatalf("Failed to initialize rabbitmq connection: %v", err)
	}

	app := application.NewApp(publisher)

	wsHandler := wb.NewWSHandler(app)

	http.HandleFunc("/ws", wsHandler.WsHandler)

	log.Println("WebSocket server starting on :8080")
	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatalf("ListenAndServe error: %v", err)
	}
}
