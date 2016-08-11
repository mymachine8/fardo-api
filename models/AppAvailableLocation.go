package models

import "gopkg.in/mgo.v2/bson"

type AppAvailableLocation struct {
	Id     bson.ObjectId `json:"id,omitempty" bson:"_id,omitempty"`
	Loc    [2]float64 `bson:"loc" json:"loc"`
	Radius int `bson:"radius" json:"radius"`
}
