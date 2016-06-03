package data

import (
	"time"
	"github.com/mymachine8/fardo-api/models"
	"gopkg.in/mgo.v2/bson"
	"github.com/mymachine8/fardo-api/common"
	"log"
)

func CreateCategory(category *models.Category) error {
	context := common.NewContext()
	defer context.Close()
	c := context.DbCollection("categories")
	obj_id := bson.NewObjectId()
	category.Id = obj_id
	category.CreatedOn = time.Now()
	err := c.Insert(&category)
	return err
}


func UpdateCategory(category *models.Category) error {
	context := common.NewContext()
	defer context.Close()
	c := context.DbCollection("categories")

	err := c.Update(bson.M{"_id": category.Id},
		bson.M{"$set": bson.M{
			"description": category.Description,
		}})
	return err
}

func GetAllCategories() []models.Category {
	context := common.NewContext()
	defer context.Close()
	c := context.DbCollection("categories")

	var categories []models.Category
	err := c.Find(nil).All(&categories)

	if(err != nil) {
		log.Println("error in getting categories");
	}

	return categories
}

