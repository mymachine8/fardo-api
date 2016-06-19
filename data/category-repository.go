package data

import (
	"time"
	"github.com/mymachine8/fardo-api/models"
	"gopkg.in/mgo.v2/bson"
	"github.com/mymachine8/fardo-api/common"
)

func CreateCategory(category *models.GroupCategory) error {
	context := common.NewContext()
	defer context.Close()
	c := context.DbCollection("group_categories")
	obj_id := bson.NewObjectId()
	category.Id = obj_id
	category.CreatedOn = time.Now()
	err := c.Insert(&category)
	return err
}


func UpdateCategory(category *models.GroupCategory) error {
	context := common.NewContext()
	defer context.Close()
	c := context.DbCollection("group_categories")

	err := c.Update(bson.M{"_id": category.Id},
		bson.M{"$set": bson.M{
			"description": category.Description,
		}})
	return err
}

func GetAllCategories() (categories []models.GroupCategory,err error) {
	context := common.NewContext()
	defer context.Close()
	c := context.DbCollection("group_categories")

	err = c.Find(nil).All(&categories)

	return categories, err
}

func CreateSubCategory(category *models.GroupSubCategory) error {
	context := common.NewContext()
	defer context.Close()
	c := context.DbCollection("group_subcategories")
	obj_id := bson.NewObjectId()
	category.Id = obj_id
	category.CreatedOn = time.Now()
	err := c.Insert(&category)
	return err
}

func UpdateSubCategory(category *models.GroupSubCategory) error {
	context := common.NewContext()
	defer context.Close()
	c := context.DbCollection("group_subcategories")

	err := c.Update(bson.M{"_id": category.Id},
		bson.M{"$set": bson.M{
			"description": category.Description,
		}})
	return err
}

func GetAllSubCategories() (categories []models.GroupSubCategory,err error) {
	context := common.NewContext()
	defer context.Close()
	c := context.DbCollection("group_subcategories")

	err = c.Find(nil).All(&categories)

	return categories, err
}

func GetSubCategories(categoryId string) (categories []models.GroupSubCategory,err error) {
	context := common.NewContext()
	defer context.Close()
	c := context.DbCollection("group_subcategories")

	err = c.Find(bson.M{"groupCategoryId": bson.ObjectIdHex(categoryId)}).All(&categories)
	return
}


