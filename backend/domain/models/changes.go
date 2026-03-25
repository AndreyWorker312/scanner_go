package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ChangeEvent represents a single detected network change stored in MongoDB.
type ChangeEvent struct {
	ID          primitive.ObjectID     `bson:"_id,omitempty"       json:"id"`
	EventID     string                 `bson:"event_id"            json:"event_id"`
	EventType   string                 `bson:"event_type"          json:"event_type"`
	Severity    string                 `bson:"severity"            json:"severity"`
	Title       string                 `bson:"title"               json:"title"`
	Description string                 `bson:"description"         json:"description"`
	Target      string                 `bson:"target"              json:"target"`
	Service     string                 `bson:"service"             json:"service"`
	Action      string                 `bson:"action"              json:"action"`
	Scanner     string                 `bson:"scanner"             json:"scanner"`
	Details     map[string]interface{} `bson:"details,omitempty"   json:"details,omitempty"`
	CreatedAt   time.Time              `bson:"created_at"          json:"created_at"`
}

