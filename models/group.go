package models

import (
	"github.com/mymachine8/fardo-api/geo"
	"time"
	"gopkg.in/mgo.v2/bson"
)

type Group struct {
	Id bson.ObjectId
	Name string
	Description string
	CategoryId int
	Radius int
	PolyLine []geo.Point
	CenterLocation geo.Point
	IsActive bool
	CreatedOn time.Time
}