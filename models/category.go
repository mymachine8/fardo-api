package models

import (
	//"github.com/mymachine8/fardo-api/bootstrap/dbconn"
)

type Category struct {
	Id int `bson:"_id" json:"id"`
	Name string `bson:"name" json:"name"`
	Description int `bson:"description" json:"description"`
}

//var CategoryCollectionName = "categories"

/*
func GetCategories() []Category {
	var categories []Category;
	mongoSession := dbconn.GetInstance();
	s := mongoSession.GetSession();
	sessionCopy := s.Copy();
	defer sessionCopy.Close();
	c := s.DB("fardo").C("categories");
	query := c.Find(nil)
	log.Printf("came here");
	err := query.All(&categories);

	if (err != nil) {
		log.Printf("Run Query Error: %s\n", err);
	}

	return categories;
}*/

