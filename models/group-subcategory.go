package models

import (
	"time"
	"gopkg.in/mgo.v2/bson"
)

type GroupSubCategory struct {
	Id bson.ObjectId `bson:"_id" json:"id"`
	Name string `bson:"name" json:"name"`
	GroupCategoryId string `bson:"groupCategoryId" json:"groupCategoryId"`
	Description int `bson:"description" json:"description"`
	CreatedOn time.Time `bson:"createdOn" json:"createdOn"`
	IsActive time.Time `bson:"isActive" json:"isActive"`
}