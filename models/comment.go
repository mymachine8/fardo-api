package models

import (
	"time"
	"gopkg.in/mgo.v2/bson"
)

type Comment struct {
	Id        bson.ObjectId `bson:"_id" json:"id"`
	PostId    bson.ObjectId `bson:"postId,omitempty" json:"postId,omitempty"`
	Content   string `bson:"content" json:"content"`
	CreatedOn time.Time `bson:"createdOn" json:"-"`
	Upvotes   int  `bson:"upvotes" json:"upvotes"`
	Downvotes int  `bson:"downvotes" json:"downvotes"`
	IsPrivate  bool `bson:"isPrivate" json:"isPrivate"`
	IsActive  bool `bson:"isActive" json:"isActive"`
}
