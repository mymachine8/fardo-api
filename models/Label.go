package models

import (
	"gopkg.in/mgo.v2/bson"
	"time"
)

type Label struct {
	Id bson.ObjectId `bson:"_id" json:"id"`
	Name string `bson:"name" json:"name"`
	Description string `bson:"description" json:"description"`
	GroupId bson.ObjectId `bson:"groupId" json:"groupId"`
	GroupName string `bson:"groupName" json:"groupName"`
	IsVerified bool `bson:"isVerified" json:"isVerified"`
	IsActive bool `bson:"isActive" json:"isActive"`
	CreatedOn time.Time `bson:"createdOn" json:"createdOn"`
}
