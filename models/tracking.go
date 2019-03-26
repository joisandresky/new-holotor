package models

import (
	"gopkg.in/mgo.v2/bson"
	"time"
)

//Tracking - models
type Tracking struct {
	ID	bson.ObjectId `json:"_id" bson:"_id,omitempty"`
	AdsID	[]int	`json:"adsId" bson:"adsId"`
	DriverID int `json:"driverId" bson:"driverId"`
	Distance	float64 `json:"distance" bson:"distance"`
	Latitude	float64	`json:"lat" bson:"lat"`
	Longitude	float64	`json:"long" bson:"long"`
	Status string	`json:"status" bson:"status"`
	Description string	`json:"description" bson:"description"`
	CreatedAt time.Time	`json:"created_at" bson:"created_at"`
}

//Trackings - models for array
type Trackings []Tracking
