package models

import (
	"github.com/mymachine8/fardo-api/geo"
)

type Group struct {
	Id int
	Name string
	CategoryId int
	Radius int
	PolyLine []geo.Point
	CenterLocation geo.Point
	IsActive bool
}