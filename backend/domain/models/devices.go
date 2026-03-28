package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// L2Device represents a unique device at the Data-Link (MAC) layer.
// Key: MAC address.
// One MAC address can have multiple associated IP addresses
// (e.g. a device with several interfaces or DHCP lease changes).
type L2Device struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	MAC         string             `bson:"mac"           json:"mac"`
	Vendor      string             `bson:"vendor"        json:"vendor"`
	IPAddresses []string           `bson:"ip_addresses"  json:"ip_addresses"`
	FirstSeen   time.Time          `bson:"first_seen"    json:"first_seen"`
	LastSeen    time.Time          `bson:"last_seen"     json:"last_seen"`
}

// L3Device represents a unique device at the Network (IP) layer.
// Key: IP address.
// References its L2 parent via the MAC field.
type L3Device struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	IP        string             `bson:"ip"            json:"ip"`
	MAC       string             `bson:"mac"           json:"mac"`
	Vendor    string             `bson:"vendor"        json:"vendor"`
	FirstSeen time.Time          `bson:"first_seen"    json:"first_seen"`
	LastSeen  time.Time          `bson:"last_seen"     json:"last_seen"`
}

