package models

import (
	"time"
	"gopkg.in/mgo.v2/bson"
)

type Affinity uint8

const (
	LocalAffinity = 0
	CategoryAffinity = 1
	MixedAffinity = 2
)

type Group struct {
	Id bson.ObjectId `bson:"_id" json:"id"`
	Name string `bson:"name" json:"name"`
	ShortName string `bson:"shortName" json:"shortName"`
	Description string `bson:"description" json:"description"`
	ImageUrl string `bson:"imageUrl" json:"imageUrl"`
	LogoData string `bson:"logoData" json:"logoData, omitempty"`
	ImageData string `bson:"-" json:"imageData, omitempty"`
	CategoryId bson.ObjectId `bson:"categoryId" json:"categoryId"`
	CategoryName string `bson:"categoryName" json:"categoryName"`
	SubCategoryId bson.ObjectId `bson:"subCategoryId" json:"subCategoryId"`
	SubCategoryName string `bson:"subCategoryName" json:"subCategoryName"`
	Radius int `bson:"radius" json:"radius"`
	Score float64 `bson:"score" json:"_"`
	CurrentPostsCount int `bson:"currentPostsCount" json:"_"`
	CurrentVotesCount int  `bson:"currentVotesCount" json:"_"`
	Loc [2]float64 `bson:"loc" json:"loc"`
	City string `bson:"city" json:"city"`
	State string `bson:"state" json:"state"`
	IsActive bool `bson:"isActive" json:"isActive"`
	CreatedOn time.Time `bson:"createdOn" json:"createdOn"`
	ModifiedOn time.Time `bson:"modifiedOn" json:"modifiedOn"`
	Affinity int `bson:"affinity" json:"-"`
}