package models

import (
	"gopkg.in/mgo.v2/bson"
)

type UserPost struct {
	Id     bson.ObjectId `json:"id" bson:"_id"`
	UserId bson.ObjectId `json:"userId" bson:"userId"`
	PostIds []bson.ObjectId `bson:"postIds" json:"postIds"`
	CommentPostIds []bson.ObjectId `bson:"commentPostIds" json:"commentPostIds"`
	CommentVotes [] UserVote `bson:"commentVotes" json:"commentVotes"`
	PostVotes [] UserVote `bson:"postVotes" json:"postVotes"`
}

