package models

import "gopkg.in/mgo.v2/bson"

type GroupLite struct {
	Id        bson.ObjectId `bson:"_id" json:"id"`
	Name      string  `bson:"name" json:"name"`
	ShortName string  `bson:"shortName" json:"shortName"`
}
