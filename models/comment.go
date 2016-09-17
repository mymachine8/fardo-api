package models

import (
	"time"
	"gopkg.in/mgo.v2/bson"
)

type Comment struct {
	Id        bson.ObjectId `bson:"_id" json:"id"`
	PostId    bson.ObjectId `bson:"postId,omitempty" json:"postId,omitempty"`
	Content   string `bson:"content" json:"content"`
	CreatedOn time.Time `bson:"createdOn" json:"createdOn"`
	ModifiedOn time.Time `bson:"modifiedOn" json:"-"`
	Upvotes   int  `bson:"upvotes" json:"upvotes"`
	Downvotes int  `bson:"downvotes" json:"downvotes"`
	VoteClicked string `bson:"-" json:"voteClicked,omitempty"`
	Replies []Reply `bson:"replies,omitempty" json:"replies,omitempty"`
	UserId    bson.ObjectId `bson:"userId" json:"userId"`
	Username string `bson:"username" json:"username"`
	IsAnonymous bool `bson:"isAnonymous" json:"isAnonymous"`
	IsPrivate  bool `bson:"isPrivate" json:"isPrivate"`
	IsActive  bool `bson:"isActive" json:"isActive"`
}
