package data

import (
	"time"
	"github.com/mymachine8/fardo-api/models"
	"gopkg.in/mgo.v2/bson"
	"github.com/mymachine8/fardo-api/common"
)


func GetFeaturedGroups(token string) (groups []models.Group, err error) {
	tokenContext := common.NewContext()
	defer tokenContext.Close()
	tokenCol := tokenContext.DbCollection("access_tokens")
	var result models.AccessToken
	err = tokenCol.Find(bson.M{"token": token}).One(&result)
	if(err != nil) {
		return
	}

	context := common.NewContext()
	defer context.Close()

	var group models.Group;
	c := context.DbCollection("groups")
	err = c.FindId(result.GroupId).One(&group);

	if(err !=nil) {
		return;
	}

	err = c.Find(bson.M{"subCategoryId": group.SubCategoryId}).Limit(30).All(&groups)
	return;
}

func CreateGroup(group models.Group) ( string , error) {
	context := common.NewContext()
	defer context.Close()
	c := context.DbCollection("groups")

	obj_id := bson.NewObjectId()
	group.Id = obj_id
	group.IsActive = true;
	group.CreatedOn = time.Now()
	err := c.Insert(&group)
	return obj_id.Hex(), err
}

func UpdateGroup(id string, group models.Group) error {
	context := common.NewContext()
	defer context.Close()
	c := context.DbCollection("groups")

	err := c.Update(bson.M{"_id": bson.ObjectIdHex(id)},
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

	err := c.Update(bson.M{"_id": bson.ObjectIdHex(id)},
		bson.M{"$set": bson.M{
			"isActive": false,
		}})
	return err
}

func GetAllGroups() (groups []models.Group, err error) {
	context := common.NewContext()
	defer context.Close()
	c := context.DbCollection("groups")

	err = c.Find(nil).All(&groups)
	return
}

func GetGroups(name string) (groups []models.Group, err error) {
	context := common.NewContext()
	defer context.Close()
	c := context.DbCollection("groups")

	name = "/" + name + "/";
	err = c.Find(bson.M{"$text": bson.M{"$search": name}}).All(&groups)
	return
}

func GetGroupById(id string) (group models.Group, err error) {
	context := common.NewContext()
	defer context.Close()
	c := context.DbCollection("groups")

	err = c.FindId(bson.ObjectIdHex(id)).One(&group)
	return
}

func CreateLabel(groupId string, label models.Label) ( string , error) {
	groupContext := common.NewContext()
	groupCol := groupContext.DbCollection("groups")
	var group models.Group
	err := groupCol.FindId(bson.ObjectIdHex(groupId)).One(&group)
	groupContext.Close()
	if(err != nil) {
		return "",err
	}
	context := common.NewContext()
	defer context.Close()
	c := context.DbCollection("labels")
	obj_id := bson.NewObjectId()
	label.Id = obj_id
	label.GroupId = bson.ObjectIdHex(groupId);
	label.CreatedOn = time.Now()
	label.GroupName = group.Name;
	err = c.Insert(&label)
	return obj_id.Hex(), err;
}

func RemoveLabel(id string) error {
	context := common.NewContext()
	defer context.Close()
	c := context.DbCollection("labels")

	err := c.Update(bson.M{"_id": bson.ObjectIdHex(id)},
		bson.M{"$set": bson.M{
			"isActive": false,
		}})
	return err;
}

func UpdateLabel(id string, label models.Label) error {
	context := common.NewContext()
	defer context.Close()
	c := context.DbCollection("labels")

	err := c.Update(bson.M{"_id": bson.ObjectIdHex(id)},
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

func GetGroupLabels(groupId string) (labels []models.Label, err error) {
	context := common.NewContext()
	defer context.Close()
	c := context.DbCollection("labels")

	err = c.Find(bson.M{"groupId": bson.ObjectIdHex(groupId)}).All(&labels)
	return
}