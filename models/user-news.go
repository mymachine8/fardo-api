package models

import "gopkg.in/mgo.v2/bson"

type UserNews struct {
	Id     bson.ObjectId `json:"id" bson:"_id"`
	UserId bson.ObjectId `json:"userId" bson:"userId"`
	NewsIds []bson.ObjectId `bson:"newsIds" json:"newsIds"`
	CommentIds []bson.ObjectId `bson:"commentPostIds" json:"commentPostIds"`
	CommentVotes [] UserVote `bson:"commentVotes" json:"commentVotes"`
	PostVotes [] UserVote `bson:"postVotes" json:"postVotes"`
}
