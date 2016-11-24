package data

import (
	"github.com/mymachine8/fardo-api/models"
	"github.com/mymachine8/fardo-api/common"
	"gopkg.in/mgo.v2/bson"
	"time"
)

func CreateLabel(groupId string, label models.Label) (string, error) {
	context := common.NewContext()
	defer context.Close()
	c := context.DbCollection("labels")
	obj_id := bson.NewObjectId()
	label.Id = obj_id
	label.CreatedOn = time.Now().UTC()
	err := c.Insert(&label)
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
			"shortName": label.ShortName,
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

func GetUserLabels(token string, groupId string, lat float64, lng float64) (labels []models.Label, err error) {
	context := common.NewContext()
	defer context.Close()
	c := context.DbCollection("labels")
	params := make(map[string]interface{})
	if (len(groupId) > 0) {
		params["groupId"] = groupId;
	}

	params["loc"] = [2]float64{lng, lat};

	options := []bson.M{}
	options = append(options, params);
	options = append(options, bson.M{"isGlobal" : true});

	err = c.Find(bson.M{"$or":options}).All(&labels)
	if (labels == nil) {
		labels = []models.Label{}
	}
	return
}

func GetAllLabels() (labels []models.Label, err error) {
	context := common.NewContext()
	defer context.Close()
	c := context.DbCollection("labels")

	err = c.Find(nil).All(&labels)
	if (labels == nil) {
		labels = []models.Label{}
	}
	return
}

func GetGroupLabels(groupId string) (labels []models.Label, err error) {
	context := common.NewContext()
	defer context.Close()
	c := context.DbCollection("labels")

	err = c.Find(bson.M{"groupId": bson.ObjectIdHex(groupId)}).All(&labels)
	if (labels == nil) {
		labels = []models.Label{}
	}
	return
}