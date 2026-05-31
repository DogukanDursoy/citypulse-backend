package models

import "go.mongodb.org/mongo-driver/bson/primitive"

// Complaint, sistemdeki şikayetlerin ana veri modelidir
type Complaint struct {
	ID         primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	Text       string             `json:"text" bson:"text"`
	Image      string             `json:"image,omitempty" bson:"image,omitempty"` // Base64 formatlı fotoğraf
	Category   string             `json:"category" bson:"category"`
	Priority   string             `json:"priority" bson:"priority"`
	Department string             `json:"department" bson:"department"`
	Status     string             `json:"status" bson:"status"` // Örn: "Pending", "Resolved"
	Lat        float64            `json:"lat" bson:"lat"`       // GPS Enlem
	Lng        float64            `json:"lng" bson:"lng"`       // GPS Boylam
}