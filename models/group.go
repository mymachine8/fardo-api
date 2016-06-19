package models

import (
	"github.com/mymachine8/fardo-api/geo"
	"time"
	"gopkg.in/mgo.v2/bson"
)

type Group struct {
	Id bson.ObjectId `bson:"_id" json:"id"`
	Name string `bson:"name" json:"name"`
	Description string `bson:"description" json:"description"`
	CategoryId bson.ObjectId `bson:"categoryId" json:"categoryId"`
	SubCategoryId bson.ObjectId `bson:"subCategoryId" json:"subCategoryId"`
	Radius int `bson:"radius" json:"radius"`
	PolyLine []geo.Point `bson:"polyline" json:"polyline"`
	Loc [2]float64 `bson:"loc" json:"loc"`
	City string `bson:"city" json:"city"`
	State string `bson:"state" json:"state"`
	IsActive bool `bson:"isActive" json:"isActive"`
	CreatedOn time.Time `bson:"createdOn" json:"createdOn"`
}