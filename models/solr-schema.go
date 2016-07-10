package models

import "gopkg.in/mgo.v2/bson"

type SolrSchema struct {
	Id        bson.ObjectId `json:"id"`
	Type      string `json:"type"`
	Name      string `json:"name"`
	ShortName string `json:"shortName,omitempty"`
	GroupName string `json:"groupName,omitempty"`
}