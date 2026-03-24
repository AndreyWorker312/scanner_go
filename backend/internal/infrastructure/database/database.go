package rabbitmq

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Database struct {
	Client   *mongo.Client
	Database *mongo.Database
}

func NewDatabase(uri, dbName string) (*Database, error) {
	log.Printf("[MongoDB] Connecting to %s, database=%s", uri, dbName)

	var (
		client  *mongo.Client
		lastErr error
	)

	for attempt := 1; attempt <= 5; attempt++ {
		log.Printf("[MongoDB] Attempt %d/5...", attempt)

		connectCtx, connectCancel := context.WithTimeout(context.Background(), 10*time.Second)
		c, err := mongo.Connect(connectCtx, options.Client().ApplyURI(uri))
		connectCancel()
		if err != nil {
			lastErr = fmt.Errorf("connect: %w", err)
			log.Printf("[MongoDB] Attempt %d/5 connect error: %v", attempt, err)
			time.Sleep(time.Duration(attempt) * 2 * time.Second)
			continue
		}

		pingCtx, pingCancel := context.WithTimeout(context.Background(), 5*time.Second)
		err = c.Ping(pingCtx, nil)
		pingCancel()
		if err != nil {
			lastErr = fmt.Errorf("ping: %w", err)
			log.Printf("[MongoDB] Attempt %d/5 ping error: %v", attempt, err)
			_ = c.Disconnect(context.Background())
			time.Sleep(time.Duration(attempt) * 2 * time.Second)
			continue
		}

		client = c
		lastErr = nil
		log.Printf("[MongoDB] Connected successfully on attempt %d", attempt)
		break
	}

	if lastErr != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB after 5 attempts: %w", lastErr)
	}

	db := client.Database(dbName)
	log.Printf("[MongoDB] Using database: %s", dbName)

	return &Database{
		Client:   client,
		Database: db,
	}, nil
}

func (d *Database) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return d.Client.Disconnect(ctx)
}

func (d *Database) ARPCollection() *mongo.Collection {
	return d.Database.Collection("arp_history")
}

func (d *Database) ICMPCollection() *mongo.Collection {
	return d.Database.Collection("icmp_history")
}

func (d *Database) NmapTcpUdpCollection() *mongo.Collection {
	return d.Database.Collection("nmap_tcp_udp_history")
}

func (d *Database) NmapOsDetectionCollection() *mongo.Collection {
	return d.Database.Collection("nmap_os_detection_history")
}

func (d *Database) NmapHostDiscoveryCollection() *mongo.Collection {
	return d.Database.Collection("nmap_host_discovery_history")
}

func (d *Database) TCPCollection() *mongo.Collection {
	return d.Database.Collection("tcp_banners")
}
