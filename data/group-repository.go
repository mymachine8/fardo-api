package data

import (
	"time"
	"github.com/mymachine8/fardo-api/models"
	"gopkg.in/mgo.v2/bson"
	"github.com/mymachine8/fardo-api/common"
	"strings"
	"encoding/base64"
	"log"
)

//Based on Affinity
//isLocal: So bring most of the local colleges to this group (Eg: Vignan's IIT, ANITS, GITAM)
//isGlobal: So bring most of the same category global colleges to this group (NIT, IIT, MEDICAL COLLEGES, NIFT)
//ismMixed: Don't know, maybe mix of local and global
func GetFeaturedGroups(token string,lat float64, lng float64) (groups []models.Group, err error) {
	tokenContext := common.NewContext()
	defer tokenContext.Close()
	tokenCol := tokenContext.DbCollection("access_tokens")
	var result models.AccessToken
	err = tokenCol.Find(bson.M{"token": token}).One(&result)
	if (err != nil) {
		return
	}

	context := common.NewContext()
	defer context.Close()

	var group models.Group;
	c := context.DbCollection("groups")
	err = c.FindId(result.GroupId).One(&group);

	if err == nil && group.Affinity == models.CategoryAffinity {
		 getFeaturedGroupsCategory(group.SubCategoryId) //50%
	} else {
		groups, err = GetNearByGroups(lat, lng); //50%
	}

	var popularGroups []models.Group;
	var adminAreaGroups []models.Group;
	groups, err = GetNearByGroups(lat, lng); //50% for local //20% for cateogry affinity
	popularGroups, err = GetGlobalPopularGroups(); //20% for local //30% for category affinity
	adminAreaGroups, err = GetAdminAreaPopularGroups("Andhra Pradesh");
	log.Print(popularGroups)
	log.Print(adminAreaGroups)
	if (groups == nil) {
		groups = []models.Group{}
	}
	return;
}

func getFeaturedGroupsCategory(subCategoryId bson.ObjectId) (groups []models.Group, err error) {
	//TODO: Get popular groups from that category
	return;
}

func GetGlobalPopularGroups() (groups []models.Group, err error) {
	//TODO: Get popular groups from the country
	return;
}

func GetAdminAreaPopularGroups(stateName string) (groups []models.Group, err error) {
	//TODO: Get popular groups from the state
	return;
}


func GetNearByGroups(lat float64, lng float64) (groups []models.Group, err error) {
	context := common.NewContext()
	defer context.Close()

	currentLatLng := [2]float64{lng, lat}
	c := context.DbCollection("groups")
	err = c.Find(bson.M{"loc":
	bson.M{"$geoWithin":
	bson.M{"$centerSphere": []interface{}{currentLatLng, 10 / 3963.2} }}}).All(&groups);
	return
}

func GetNearByPopularGroups(lat float64, lng float64) (groups []models.Group, err error) {
	//TODO: Get popular groups from the nearby groups
	context := common.NewContext()
	defer context.Close()

	currentLatLng := [2]float64{lng, lat}
	c := context.DbCollection("groups")
	err = c.Find(bson.M{"loc":
	bson.M{"$geoWithin":
	bson.M{"$centerSphere": []interface{}{currentLatLng, 10 / 3963.2} }}}).All(&groups);
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
	err := c.Insert(group)
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

func GetAllGroups() (groups []models.Group, err error) {
	context := common.NewContext()
	defer context.Close()
	c := context.DbCollection("groups")

	err = c.Find(nil).Sort("-createdOn").All(&groups);
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