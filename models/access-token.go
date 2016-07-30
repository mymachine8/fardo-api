package models

import (
	"gopkg.in/mgo.v2/bson"
	"time"
)

type AccessToken  struct {
	Id        bson.ObjectId `bson:"_id" json:"id"`
	UserId    bson.ObjectId `bson:"userId" json:"userId"`
	GroupId   bson.ObjectId `bson:"groupId,omitempty" json:"groupId,omitempty"`
	Token     string `bson:"token" json:"token"`
	FcmToken     string `bson:"fcmToken" json:"fcmToken"`
	CreatedOn time.Time `bson:"createdOn" json:"-"`
	ModifiedOn time.Time `bson:"modifiedOn" json:"-"`
}

