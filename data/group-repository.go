package data

import (
	"time"
	"github.com/mymachine8/fardo-api/models"
	"gopkg.in/mgo.v2/bson"
	"github.com/mymachine8/fardo-api/common"
	"strings"
	"encoding/base64"
)

func GetPopularGroups(lat float64, lng float64) (groups []models.GroupLite, err error) {
	//TODO: Get based on number of posts in the last 2 days
	//TODO: Maintain Global popular groups in a buffer

	var localGroups []models.GroupLite;
	var globalGroups []models.GroupLite;

	localGroups, err = GetNearByPopularGroups(lat,lng);
	globalGroups, err = GetGlobalPopularGroups();

	var i int;
	for i=0;i<3 && i < len(localGroups);i++ {
		groups = append(groups, localGroups[i]);
	}

	remaining := 3 - i;

	for i = 0; i< 3 + remaining && i < len(globalGroups); i++ {
		if(!idInGroups(globalGroups[i].Id.Hex(), groups)) {
			groups = append(groups, globalGroups[i]);
		}
	}


	if (groups == nil) {
		groups = []models.GroupLite{}
	}
	return;
}

func idInGroups(id string, list []models.GroupLite) bool {
	for _, b := range list {
		if b.Id.Hex() == id {
			return true
		}
	}
	return false
}

func GetGlobalPopularGroups() (groups []models.GroupLite, err error) {
	context := common.NewContext()
	defer context.Close()
	c := context.DbCollection("groups")
	err = c.Find(nil).Sort("-score").Limit(30).All(&groups)
	return;
}

func GetNearByPopularGroups(lat float64, lng float64) (groups []models.GroupLite, err error) {
	context := common.NewContext()
	defer context.Close()

	currentLatLng := [2]float64{lng, lat}
	c := context.DbCollection("groups")
	err = c.Find(bson.M{"loc":
	bson.M{"$geoWithin":
	bson.M{"$centerSphere": []interface{}{currentLatLng, 30 / 3963.2} }}}).Select(bson.M{"name": 1, "shortName": 1}).Limit(6).All(&groups);
	return
}


func GetNearByGroups(lat float64, lng float64) (groups []models.Group, err error) {
	context := common.NewContext()
	defer context.Close()

	currentLatLng := [2]float64{lng, lat}
	c := context.DbCollection("groups")
	err = c.Find(bson.M{"loc":
	bson.M{"$geoWithin":
	bson.M{"$centerSphere": []interface{}{currentLatLng, 30 / 3963.2} }}}).All(&groups);
	return
}

func CreateGroup(group models.Group) (string, error) {
	context := common.NewContext()
	defer context.Close()
	c := context.DbCollection("groups")
	group.Id = bson.NewObjectId()
	group.IsActive = true;
	group.CreatedOn = time.Now().UTC()

	if (len(group.ImageData) > 0 ) {
		fileName := "group_image_" + group.Id.Hex();
		imageReader := strings.NewReader(group.ImageData);

		dec := base64.NewDecoder(base64.StdEncoding, imageReader);

		res, err := common.SendItemToCloudStorage(common.GroupImage, fileName, dec);

		if (err != nil) {
			return "", models.FardoError{"Insert Group Image Error: " + err.Error()}
		}

		group.ImageUrl = res;

	}

	category, err := GetCategory(group.CategoryId.Hex());

	if(err != nil) {
		return "", err
	}

	group.CategoryName = category.Name

	err = c.Insert(&group)
	return group.Id.Hex(), err
}

func UpdateGroup(id string, group models.Group) error {
	context := common.NewContext()
	defer context.Close()
	c := context.DbCollection("groups")

	err := c.Update(bson.M{"_id": bson.ObjectIdHex(id)},
		group);
	return err
}

func SuspendGroup(id string) error {
	context := common.NewContext()
	defer context.Close()
	c := context.DbCollection("groups")

	err := c.Update(bson.M{"_id": bson.ObjectIdHex(id)},
		bson.M{"$set": bson.M{
			"isActive": false,
			"modifiedOn": time.Now().UTC(),
		}})
	return err
}

func UpdateGroupLogo(id string, groupLogo string) error {
	context := common.NewContext()
	defer context.Close()
	c := context.DbCollection("groups")

	err := c.Update(bson.M{"_id": bson.ObjectIdHex(id)},
		bson.M{"$set": bson.M{"logoData": groupLogo}});
	return err
}

func UpdateGroupImage(id string, imageData string) error {
	context := common.NewContext()
	defer context.Close()
	c := context.DbCollection("groups")

	imageId := bson.NewObjectId()
	fileName := "group_image_" + imageId.Hex();
	imageReader := strings.NewReader(imageData);

	dec := base64.NewDecoder(base64.StdEncoding, imageReader);

	res, err := common.SendItemToCloudStorage(common.GroupImage, fileName, dec);

	if (err != nil) {
		return models.FardoError{"Insert Post Image Error: " + err.Error()}
	}

	err = c.Update(bson.M{"_id": bson.ObjectIdHex(id)},
		bson.M{"$set": bson.M{"imageUrl": res}});
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

func GetGroups(page int,groupParams models.Group) (groups []models.Group, err error) {
	context := common.NewContext()
	defer context.Close()
	c := context.DbCollection("groups")

	skip := page*20;
	params := make(map[string]interface{})
	if(len(groupParams.Name) > 0) {
		params["name"] = bson.RegEx{Pattern: groupParams.Name, Options: "i"};
	}
	if(len(groupParams.City) > 0) {
		params["city"] = bson.RegEx{Pattern: groupParams.City, Options: "i"};
	}
	if(len(groupParams.State) > 0) {
		params["state"] = bson.RegEx{Pattern: groupParams.State, Options: "i"};
	}

	err = c.Find(params).Sort("-createdOn").Skip(skip).Limit(20).All(&groups);
	if (groups == nil) {
		groups = []models.Group{}
	}
	return
}

func GetAllGroups(name string) (groups []models.Group, err error) {
	context := common.NewContext()
	defer context.Close()
	c := context.DbCollection("groups")

	name = "/" + name + "/";
	err = c.Find(bson.M{"$text": bson.M{"$search": name}}).All(&groups)

	if (groups == nil) {
		groups = []models.Group{}
	}
	return
}

func GetGroupById(id string) (group models.Group, err error) {
	context := common.NewContext()
	defer context.Close()
	c := context.DbCollection("groups")

	err = c.FindId(bson.ObjectIdHex(id)).One(&group)
	return
}

func CreateLabel(groupId string, label models.Label) (string, error) {
	groupContext := common.NewContext()
	groupCol := groupContext.DbCollection("groups")
	var group models.Group
	err := groupCol.FindId(bson.ObjectIdHex(groupId)).One(&group)
	groupContext.Close()
	if (err != nil) {
		return "", err
	}
	context := common.NewContext()
	defer context.Close()
	c := context.DbCollection("labels")
	obj_id := bson.NewObjectId()
	label.Id = obj_id
	label.GroupId = bson.ObjectIdHex(groupId);
	label.CreatedOn = time.Now().UTC()
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
