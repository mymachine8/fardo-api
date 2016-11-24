package models

import (
	"gopkg.in/mgo.v2/bson"
	"time"
)

type Label struct {
	Id           bson.ObjectId `bson:"_id" json:"id"`
	Name         string `bson:"name" json:"name"`
	ShortName    string `bson:"shortName" json:"shortName"`
	LabelIconUrl string  `bson:"labelIconUrl" json:"labelIconUrl"`
	Description  string `bson:"description" json:"description"`
	IsVerified   bool `bson:"isVerified" json:"isVerified"`
	IsActive     bool `bson:"isActive" json:"isActive"`
	Loc          [2]float64 `bson:"loc" json:"loc"`
	IsGlobal     bool `bson:"isGlobal" json:"isGlobal"`
	CreatedOn    time.Time `bson:"createdOn" json:"createdOn"`
	ModifiedOn   time.Time `bson:"modifiedOn" json:"modifiedOn"`
	SuggestedBy  bson.ObjectId `bson:"suggestedBy" json:"suggestedBy"`
}
