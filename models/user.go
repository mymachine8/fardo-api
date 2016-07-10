package models

import (
	"gopkg.in/mgo.v2/bson"
	"time"
)

type User struct {
	Id bson.ObjectId `json:"id,omitempty" bson:"_id,omitempty"`
	Username string `bson:"username,omitempty" json:"username,omitempty"`
	Password string `bson:"password,omitempty" json:"password,omitempty"`
	HashPassword []byte `bson:"hashPassword,omitempty" json:"hashPassword, omitempty"`
	Imei string `json:"imei,omitempty" bson:"imei,omitempty"`
	Status string `json:"status" bson:"status"`
	Score int `json:"score" bson:"score"`
	LastKnowLocation [2]float64 `bson:"lastKnownLocation" json:"lastKnownLocation"`
	Phone string `json:"phone,omitempty" bson:"phone,omitempty"`
	GroupId bson.ObjectId `json:"groupId,omitempty" bson:"groupId,omitempty"`
	IsGroupLocked bool `json:"isGroupLocked,omitempty" bson:"isGroupLocked,omitempty"`
	IsActive bool `bson:"isActive" json:"isActive"`
	Role string `bson:"role"`
	CreatedOn time.Time `bson:"createdOn" json:"-"`
	ModifiedOn time.Time `bson:"modifiedOn" json:"-"`
}
