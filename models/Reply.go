package models

import (
	"time"
	"gopkg.in/mgo.v2/bson"
)

type Reply struct {
	Content   string `bson:"content" json:"content"`
	CreatedOn time.Time `bson:"createdOn" json:"-"`
	UserId    bson.ObjectId `bson:"userId" json:"-"`
}
