package models

import "gopkg.in/mgo.v2/bson"

type UserVote struct {
	Id bson.ObjectId `json:"id" bson:"_id"`
	IsUpvote bool `json:"isUpvote" bson:"isUpvote"`
}
