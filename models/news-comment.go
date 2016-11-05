package models

import (
	"gopkg.in/mgo.v2/bson"
	"time"
)

type NewsComment struct {
	Id        bson.ObjectId `bson:"_id" json:"id"`
	NewsId    bson.ObjectId `bson:"newsId,omitempty" json:"newsId,omitempty"`
	Content   string `bson:"content" json:"content"`
	CreatedOn time.Time `bson:"createdOn" json:"createdOn"`
	ModifiedOn time.Time `bson:"modifiedOn" json:"-"`
	Upvotes   int  `bson:"upvotes" json:"upvotes"`
	Downvotes int  `bson:"downvotes" json:"downvotes"`
	VoteClicked string `bson:"-" json:"voteClicked,omitempty"`
	UserId    bson.ObjectId `bson:"userId" json:"userId"`
	Username string `bson:"username" json:"username"`
	IsPrivate  bool `bson:"isPrivate" json:"isPrivate"`
	IsActive  bool `bson:"isActive" json:"isActive"`
}
