package models

import (
	//"github.com/mymachine8/fardo-api/bootstrap/dbconn"
)
import (
	"time"
	"gopkg.in/mgo.v2/bson"
)

type GroupCategory struct {
	Id bson.ObjectId `bson:"_id" json:"id"`
	Name string `bson:"name" json:"name"`
	Description int `bson:"description" json:"description"`
	CreatedOn time.Time `bson:"createdOn" json:"createdOn"`
	IsActive time.Time `bson:"isActive" json:"isActive"`
}
