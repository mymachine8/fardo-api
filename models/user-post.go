package models

import (
	"gopkg.in/mgo.v2/bson"
)

type UserPost struct {
	Id     bson.ObjectId `json:"id,omitempty" bson:"_id,omitempty"`
	UserId bson.ObjectId `json:"userId,omitempty" bson:"userId,omitempty"`
	PostIds []bson.ObjectId `bson:"postIds" json:"postIds"`
	CommentPostIds []bson.ObjectId `bson:"commentPostIds" json:"commentPostIds"`
}

