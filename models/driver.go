package models

import (
	"time"

	"gopkg.in/mgo.v2/bson"
)

// Driver - model
type Driver struct {
	ID        bson.ObjectId `json:"id" bson:"_id,omitempty"`
	DriverID  string        `json:"driver_id" bson:"driver_id"`
	Daily     float64       `json:"daily" bson:"daily"`
	Annualy   float64       `json:"annualy" bson:"annualy"`
	Total     float64       `json:"total" bson:"total"`
	UpdatedAt time.Time     `json:"updated_at" bson:"updated_at"`
}

// Drivers - model
type Drivers []Driver
