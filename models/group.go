package models

import (
	"time"
	"gopkg.in/mgo.v2/bson"
)

type Group struct {
	Id bson.ObjectId `bson:"_id" json:"id"`
	Name string `bson:"name" json:"name"`
	ShortName string `bson:"shortName" json:"shortName"`
	Description string `bson:"description" json:"description"`
	ImageUrl string `bson:"imageUrl" json:"imageUrl"`
	LogoData string `bson:"-" json:"logoData,omitempty"`
	LogoUrl string `bson:"logoUrl" json:"logoUrl,omitempty"`
	ImageData string `bson:"-" json:"imageData,omitempty"`
	CategoryId bson.ObjectId `bson:"categoryId" json:"categoryId"`
	CategoryName string `bson:"categoryName" json:"categoryName"`
	SubCategoryId bson.ObjectId `bson:"subCategoryId" json:"subCategoryId"`
	SubCategoryName string `bson:"subCategoryName" json:"subCategoryName"`
	Radius int `bson:"radius" json:"radius"`
	PostsCount int `bson:"postsCount" json:"postsCount"`
	Scores []int `bson:"scores" json:"-"`
	IsChildGroup bool `bson:"isChildGroup" json:"isChildGroup"`
	ParentGroupId bson.ObjectId `bson:"parentGroupId,omitempty" json:"parentGroupId,omitempty"`
	ScoreLastUpdated time.Time `bson:"scoreLastUpdated" json:"scoreLastUpdated"`
	TrendingScore int `bson:"trendingScore" json:"trendingScore"`
	Loc [2]float64 `bson:"loc" json:"loc"`
	City string `bson:"city" json:"city"`
	State string `bson:"state" json:"state"`
	IsActive bool `bson:"isActive" json:"isActive"`
	CreatedOn time.Time `bson:"createdOn" json:"createdOn"`
	ModifiedOn time.Time `bson:"modifiedOn" json:"modifiedOn"`
}