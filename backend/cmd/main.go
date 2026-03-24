package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"sync/atomic"
	"time"
	"unsafe"

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

// appHolder holds a pointer to *application.App atomically so it can be
// swapped in once RabbitMQ is available without restarting the HTTP server.
var appHolder unsafe.Pointer // stores *application.App

func loadApp() *application.App {
	p := atomic.LoadPointer(&appHolder)
	if p == nil {
		return nil
	}
	return (*application.App)(p)
}

func storeApp(a *application.App) {
	atomic.StorePointer(&appHolder, unsafe.Pointer(a))
}

func main() {
	rabbitMQURL := getEnv("RABBITMQ_URL", "amqp://guest:guest@rabbitmq:5672/")
	mongoURI    := getEnv("MONGODB_URI",   "mongodb://mongodb:27017")
	mongoDB     := getEnv("MONGODB_DATABASE", "network_scanner")

	// ── MongoDB ──────────────────────────────────────────────────────────────
	log.Println("[Main] Connecting to MongoDB…")
	db, err := database.NewDatabase(mongoURI, mongoDB)
	if err != nil {
		log.Fatalf("[Main] Failed to connect to MongoDB: %v", err)
	}
	defer db.Close()
	repo := database.NewRepository(db)
	log.Println("[Main] MongoDB connected")

	// ── HTTP routes that do NOT need RabbitMQ ────────────────────────────────
	historyHandler := rest.NewHistoryHandler(repo)
	searchHandler  := rest.NewSearchHandler(repo, nil) // app set later

	http.HandleFunc("/api/history/arp",    historyHandler.GetARPHistory)
	http.HandleFunc("/api/history/icmp",   historyHandler.GetICMPHistory)
	http.HandleFunc("/api/history/nmap",   historyHandler.GetNmapHistory)
	http.HandleFunc("/api/history/tcp",    historyHandler.GetTCPHistory)

	http.HandleFunc("/api/history/icmp/by-id",              searchHandler.GetICMPHistoryByID)
	http.HandleFunc("/api/history/nmap/tcp_udp/by-id",      searchHandler.GetNmapTcpUdpHistoryByID)
	http.HandleFunc("/api/history/nmap/os_detection/by-id", searchHandler.GetNmapOsDetectionHistoryByID)
	http.HandleFunc("/api/history/nmap/host_discovery/by-id",searchHandler.GetNmapHostDiscoveryHistoryByID)
	http.HandleFunc("/api/history/arp/by-id",               searchHandler.GetARPHistoryByID)
	http.HandleFunc("/api/history/tcp/by-id",               searchHandler.GetTCPHistoryByID)

	http.HandleFunc("/api/search/icmp", searchHandler.SearchICMP)
	http.HandleFunc("/api/search/nmap", searchHandler.SearchNmap)
	http.HandleFunc("/api/search/arp",  searchHandler.SearchARP)
	http.HandleFunc("/api/search/tcp",  searchHandler.SearchTCP)

	http.HandleFunc("/api/history/arp/delete",  historyHandler.DeleteARPHistory)
	http.HandleFunc("/api/history/icmp/delete", historyHandler.DeleteICMPHistory)
	http.HandleFunc("/api/history/nmap/delete", historyHandler.DeleteNmapHistory)
	http.HandleFunc("/api/history/tcp/delete",  historyHandler.DeleteTCPHistory)

	// ── /ws — proxied through appHolder; returns 503 while RabbitMQ not ready ─
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		app := loadApp()
		if app == nil {
			log.Printf("[WS] RabbitMQ not ready yet — returning 503")
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusServiceUnavailable)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "backend not ready, RabbitMQ connecting…",
			})
			return
		}
		wb.NewWSHandler(app).WsHandler(w, r)
	})

	// ── /health — always responds 200 ────────────────────────────────────────
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		ready := loadApp() != nil
		w.Header().Set("Content-Type", "application/json")
		if ready {
			json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
		} else {
			w.WriteHeader(http.StatusServiceUnavailable)
			json.NewEncoder(w).Encode(map[string]string{"status": "starting", "detail": "waiting for rabbitmq"})
		}
	})

	// ── Start HTTP server IMMEDIATELY ────────────────────────────────────────
	log.Println("[Main] HTTP server starting on :8080 (RabbitMQ connecting in background)")
	go func() {
		if err := http.ListenAndServe(":8080", nil); err != nil {
			log.Fatalf("[Main] ListenAndServe error: %v", err)
		}
	}()

	// ── Connect to RabbitMQ in background ────────────────────────────────────
	log.Println("[Main] Starting RabbitMQ connection in background…")
	publisher, err := rabbitmq.GetRPCconnection(rabbitMQURL)
	if err != nil {
		log.Printf("[Main] WARNING: RabbitMQ connection failed after all retries: %v", err)
		log.Printf("[Main] Scan endpoints will be unavailable. Restarting container…")
		// Exit so Docker restarts us; HTTP history endpoints remain up until then
		os.Exit(1)
	}

	// ── Wire up the full app once RabbitMQ is available ──────────────────────
	app := application.NewApp(publisher, repo)
	storeApp(app)
	// also update searchHandler's app reference
	searchHandler.SetApp(app)
	log.Println("[Main] RabbitMQ connected — WebSocket scan endpoint is now active")

	// Keep main goroutine alive
	select {}
}

// keep process alive by sleeping forever after http server launched
func keepAlive() {
	for {
		time.Sleep(24 * time.Hour)
	}
}
