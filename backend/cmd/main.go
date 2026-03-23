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

	repo := database.NewRepository(db)

	app := application.NewApp(publisher, repo)

	wsHandler := wb.NewWSHandler(app)
	historyHandler := rest.NewHistoryHandler(repo)
	searchHandler := rest.NewSearchHandler(repo, app)

	http.HandleFunc("/ws", wsHandler.WsHandler)

	http.HandleFunc("/api/history/arp", historyHandler.GetARPHistory)
	http.HandleFunc("/api/history/icmp", historyHandler.GetICMPHistory)
	http.HandleFunc("/api/history/nmap", historyHandler.GetNmapHistory)
	http.HandleFunc("/api/history/tcp", historyHandler.GetTCPHistory)

	http.HandleFunc("/api/history/icmp/by-id", searchHandler.GetICMPHistoryByID)
	http.HandleFunc("/api/history/nmap/tcp_udp/by-id", searchHandler.GetNmapTcpUdpHistoryByID)
	http.HandleFunc("/api/history/nmap/os_detection/by-id", searchHandler.GetNmapOsDetectionHistoryByID)
	http.HandleFunc("/api/history/nmap/host_discovery/by-id", searchHandler.GetNmapHostDiscoveryHistoryByID)
	http.HandleFunc("/api/history/arp/by-id", searchHandler.GetARPHistoryByID)
	http.HandleFunc("/api/history/tcp/by-id", searchHandler.GetTCPHistoryByID)

	http.HandleFunc("/api/search/icmp", searchHandler.SearchICMP)
	http.HandleFunc("/api/search/nmap", searchHandler.SearchNmap)
	http.HandleFunc("/api/search/arp", searchHandler.SearchARP)
	http.HandleFunc("/api/search/tcp", searchHandler.SearchTCP)

	http.HandleFunc("/api/history/arp/delete", historyHandler.DeleteARPHistory)
	http.HandleFunc("/api/history/icmp/delete", historyHandler.DeleteICMPHistory)
	http.HandleFunc("/api/history/nmap/delete", historyHandler.DeleteNmapHistory)
	http.HandleFunc("/api/history/tcp/delete", historyHandler.DeleteTCPHistory)

	log.Println("Server starting on :8080")
	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatalf("ListenAndServe error: %v", err)
	}
}
