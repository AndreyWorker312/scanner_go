package main

import (
	"log"
	"net/http"
	"os"

	"backend/internal/application"
	database "backend/internal/infrastructure/database"
	rabbitmq "backend/internal/infrastructure/messaging"
	rest "backend/internal/presentation/http"
	wb "backend/internal/presentation/websocket"
)

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func main() {
	rabbitMQURL := getEnv("RABBITMQ_URL", "amqp://guest:guest@rabbitmq:5672/")
	publisher, err := rabbitmq.GetRPCconnection(rabbitMQURL)
	if err != nil {
		log.Fatalf("Failed to initialize rabbitmq connection: %v", err)
	}

	mongoURI := getEnv("MONGODB_URI", "mongodb://mongodb:27017")
	mongoDB := getEnv("MONGODB_DATABASE", "network_scanner")
	db, err := database.NewDatabase(mongoURI, mongoDB)
	if err != nil {
		log.Fatalf("Failed to initialize MongoDB connection: %v", err)
	}
	defer db.Close()

	// Initialize Repository
	repo := database.NewRepository(db)

	// Initialize App with both publisher and repository
	app := application.NewApp(publisher, repo)

	// Set up callback for RPC responses to save to database
	publisher.SetResponseCallback(app.ProcessResponse)

	wsHandler := wb.NewWSHandler(app)
	historyHandler := rest.NewHistoryHandler(repo)

	// WebSocket endpoint
	http.HandleFunc("/ws", wsHandler.WsHandler)

	// REST API endpoints for history
	http.HandleFunc("/api/history/arp", historyHandler.GetARPHistory)
	http.HandleFunc("/api/history/icmp", historyHandler.GetICMPHistory)
	http.HandleFunc("/api/history/nmap", historyHandler.GetNmapHistory)
	http.HandleFunc("/api/history/tcp", historyHandler.GetTCPHistory)

	// DELETE endpoints for history
	http.HandleFunc("/api/history/arp/delete", historyHandler.DeleteARPHistory)
	http.HandleFunc("/api/history/icmp/delete", historyHandler.DeleteICMPHistory)
	http.HandleFunc("/api/history/nmap/delete", historyHandler.DeleteNmapHistory)
	http.HandleFunc("/api/history/tcp/delete", historyHandler.DeleteTCPHistory)

	// Static files
	http.Handle("/", http.FileServer(http.Dir("./cmd/public")))

	log.Println("Server starting on :8080")
	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatalf("ListenAndServe error: %v", err)
	}
}
