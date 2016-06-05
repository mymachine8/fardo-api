package models

import (
	"time"
	"gopkg.in/mgo.v2/bson"
)

type PostLite struct {
	//UserId bson.ObjectId
	Id        bson.ObjectId `bson:"_id" json:"id"`
	Loc       [2]float64  `bson:"loc" json:"loc"`
	CreatedOn time.Time `bson:"createdOn" json:"-"`
	GroupId   bson.ObjectId `bson:"groupId,omitempty" json:"groupId,omitempty"`
	LabelId   bson.ObjectId `bson:"labelId,omitempty" json:"labelId,omitempty"`
}

