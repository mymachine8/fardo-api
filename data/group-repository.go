package data

import (
	"time"
	"github.com/mymachine8/fardo-api/models"
	"gopkg.in/mgo.v2/bson"
	"github.com/mymachine8/fardo-api/common"
)

func CreateGroup(group *models.Group) error {
	context := common.NewContext()
	defer context.Close()
	c := context.DbCollection("groups")

	obj_id := bson.NewObjectId()
	group.Id = obj_id
	group.CreatedOn = time.Now()
	err := c.Insert(&group)
	return err
}

func UpdateGroup(group *models.Group) error {
	context := common.NewContext()
	defer context.Close()
	c := context.DbCollection("groups")

	err := c.Update(bson.M{"_id": group.Id},
		bson.M{"$set": bson.M{
			"name": group.Name,
			"description": group.Description,
		}})
	return err
}

func RemoveGroup(id string) error {
	context := common.NewContext()
	defer context.Close()
	c := context.DbCollection("groups")

	err := c.Update(bson.M{"_id": id},
		bson.M{"$set": bson.M{
			"isActive": false,
		}})
	return err
}

func GetAllGroups() (groups []models.Group, err error) {
	context := common.NewContext()
	defer context.Close()
	c := context.DbCollection("groups")

	var groups []models.Group
	err = c.Find(nil).All(&groups)
	return
}

func GetGroupById(id string) (group models.Group, err error) {
	context := common.NewContext()
	defer context.Close()
	c := context.DbCollection("groups")

	err = c.FindId(bson.ObjectIdHex(id)).One(&group)
	return
}

func CreateLabel(label *models.Label) error {
	context := common.NewContext()
	defer context.Close()
	c := context.DbCollection("labels")

	obj_id := bson.NewObjectId()
	label.Id = obj_id
	label.CreatedOn = time.Now()
	err := c.Insert(&label)
	return err;
}

func DeleteLabel(id string) (label models.Label, err error) {
	context := common.NewContext()
	defer context.Close()
	c := context.DbCollection("labels")

	err = c.Update(bson.M{"_id": id},
		bson.M{"$set": bson.M{
			"isActive": false,
		}})
	return
}

func UpdateLabel(label *models.Label) error {
	context := common.NewContext()
	defer context.Close()
	c := context.DbCollection("labels")

	err := c.Update(bson.M{"_id": label.Id},
		bson.M{"$set": bson.M{
			"name": label.Name,
			"description": label.Description,
		}})
	return err
}

func GetLabelById(id string) (label models.Label, err error) {
	context := common.NewContext()
	defer context.Close()
	c := context.DbCollection("labels")

	err = c.FindId(bson.ObjectIdHex(id)).One(&label)
	return
}

func GetAllLabels() (labels []models.Label, err error) {
	context := common.NewContext()
	defer context.Close()
	c := context.DbCollection("labels")

	err = c.Find(nil).All(&labels)
	return
}