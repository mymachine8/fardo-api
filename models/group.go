package models

import (
	"gopkg.in/mgo.v2/bson"
	"github.com/mymachine8/fardo-api/geo"
)

type Group struct {
	name string
	categoryId []bson.ObjectId
	polyLine []geo.Point
	centerLocation geo.Point
}