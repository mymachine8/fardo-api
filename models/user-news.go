package models

import "gopkg.in/mgo.v2/bson"

type UserNews struct {
	Id     bson.ObjectId `json:"id" bson:"_id"`
	UserId bson.ObjectId `json:"userId" bson:"userId"`
	NewsIds []bson.ObjectId `bson:"newsIds" json:"newsIds"`
	CommentNewsIds []bson.ObjectId `bson:"CommentNewsIds" json:"CommentNewsIds"`
	CommentVotes [] UserVote `bson:"commentVotes" json:"commentVotes"`
	Votes [] UserVote `bson:"votes" json:"votes"`
}
