package models

import (
	"gopkg.in/mgo.v2/bson"
	"time"
)

type Label struct {
	Id bson.ObjectId
	Name string
	Description string
	GroupId int
	IsVerified bool
	IsActive bool
	CreatedOn time.Time
}
