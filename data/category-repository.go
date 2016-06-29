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
	category.Id = bson.NewObjectId()
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

func CreateSubCategories(categories []models.GroupSubCategory) (err error) {
	context := common.NewContext()
	defer context.Close()
	c := context.DbCollection("group_subcategories")

	for i := 0; i < len(categories); i++ {
		category := &categories[i]
		category.Id = bson.NewObjectId();
		err = c.Insert(&category)
		if(err!=nil) {
			return
		}
	}

	return
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

	err = c.Find(bson.M{"categoryId": bson.ObjectIdHex(categoryId)}).All(&categories)
	return
}


