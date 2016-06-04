package models

import (
	"gopkg.in/mgo.v2/bson"
)

type User struct {
	Id bson.ObjectId `json:"id,omitempty" bson:"_id,omitempty"`
	Username string `bson:"username,omitempty" json:"username,omitempty"`
	Password string `bson:"password,omitempty" json:"password,omitempty"`
	HashPassword []byte `bson:"hashPassword,omitempty" json:"hashPassword, omitempty"`
	Imei string `json:"imei,omitempty" bson:"imei,omitempty"`
	Status string `json:"status" bson:"status"`
	Score int `json:"score" bson:"score"`
	LastKnowLocation []float64 `bson:"lastKnownLocation" json:"lastKnownLocation"`
	Phone string `json:"phone,omitempty" bson:"phone,omitempty"`
	GroupIds []bson.ObjectId `json:"groupIds,omitempty" bson:"groupIds,omitempty"`
	IsActive bool `bson:"isActive" json:"isActive"`
	Role string `bson:"role"`
}
