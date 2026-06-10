package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Complaint, sistemdeki şikayetlerin tek ve ana veri modelidir
type Complaint struct {
	ID           primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	UserText     string             `json:"user_text" bson:"user_text"`
	Category     string             `json:"category" bson:"category"`
	Priority     string             `json:"priority" bson:"priority"`
	Department   string             `json:"department" bson:"department"`
	Status       string             `json:"status" bson:"status"`
	TrackingCode string             `json:"tracking_code" bson:"tracking_code"`
	Image        string             `json:"image,omitempty" bson:"image,omitempty"`
	Lat          float64            `json:"lat" bson:"lat"`
	Lng          float64            `json:"lng" bson:"lng"`
	CreatedAt    time.Time          `json:"created_at" bson:"created_at"`
}
