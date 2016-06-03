package models

import (
	"gopkg.in/mgo.v2/bson"
)

type User struct {
	Id bson.ObjectId `json:"id,omitempty" bson:"_id,omitempty"`
	Imei string `json:"imei,omitempty" bson:"imei,omitempty"`
	Status string `json:"status" bson:"status"`
	Score string `json:"score" bson:"score"`
	LastKnowLoc []float64 `bson:"last_known_loc"`
	Phone string `json:"phone,omitempty" bson:"phone,omitempty"`
	GroupIds []bson.ObjectId `json:"groupIds,omitempty" bson:"groups,omitempty"`
	CountryCode string `bson:"groups,omitempty"`
	IsActive bool
}
