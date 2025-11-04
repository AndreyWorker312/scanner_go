package rabbitmq

import (
	"context"
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
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}

	// Test connection
	err = client.Ping(ctx, nil)
	if err != nil {
		return nil, err
	}

	db := client.Database(dbName)
	log.Printf("Connected to MongoDB database: %s", dbName)

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

// Get collections
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
