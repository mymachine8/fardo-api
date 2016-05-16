package models

import (
	"gopkg.in/mgo.v2/bson"
)

type User struct {
	id bson.ObjectId `json:"id,omitempty" bson:"_id,omitempty"`
	imei string `json:"imei,omitempty" bson:"imei,omitempty"`
	status string `json:"status" bson:"status"`
	score string `json:"score" bson:"score"`
	lastKnowLoc []float64 `bson:"last_known_loc"`
	phone string `json:"phone,omitempty" bson:"phone,omitempty"`
	groupIds []bson.ObjectId `json:"groupIds,omitempty" bson:"groups,omitempty"`
	countryCode string `bson:"groups,omitempty"`
}
