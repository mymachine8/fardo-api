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
	tokenCol := tokenContext.DbCollection("users")
	var result models.User
	err = tokenCol.Find(bson.M{"token": token}).One(&result)
	if (err != nil) {
		return
	}

	context := common.NewContext()
	defer context.Close()

	var group models.Group;
	c := context.DbCollection("groups")
	err = c.FindId(result.GroupId).One(&group);

	var featuredGroups []models.Group
	if err == nil && group.Affinity == models.CategoryAffinity {
		 featuredGroups, err = getFeaturedGroupsCategory(group.SubCategoryId) //50%
	}

	var popularGroups []models.Group;
	var adminAreaGroups []models.Group;

	popularGroups, err = GetGlobalPopularGroups(); //20% for local //30% for category affinity
	adminAreaGroups, err = GetAdminAreaPopularGroups(lat,lng);
	log.Print(popularGroups)
	log.Print(adminAreaGroups)
	log.Print(featuredGroups)
	if (groups == nil) {
		groups = []models.Group{}
	}
	return;
}

func getFeaturedGroupsCategory(subCategoryId bson.ObjectId) (groups []models.Group, err error) {
	context := common.NewContext()
	defer context.Close()
	c := context.DbCollection("groups")
	err = c.Find(bson.M{"subCategoryId": subCategoryId}).Sort("-score").Limit(30).All(&groups)
	return;
}

func GetGlobalPopularGroups() (groups []models.Group, err error) {
	context := common.NewContext()
	defer context.Close()
	c := context.DbCollection("groups")
	err = c.Find(nil).Sort("-score").Limit(30).All(&groups)
	return;
}

func GetAdminAreaPopularGroups(lat float64, lng float64) (groups []models.Group, err error) {
	context := common.NewContext()
	defer context.Close()

	currentLatLng := [2]float64{lng, lat}
	c := context.DbCollection("groups")
	err = c.Find(bson.M{"loc":
	bson.M{"$geoWithin":
	bson.M{"$centerSphere": []interface{}{currentLatLng, 30 / 3963.2} }}}).Sort("-score").All(&groups);
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


func GetNearByGroupsScore(lat float64, lng float64) ( interface{}, error) {

	context := common.NewContext()
	defer context.Close()

	var groupScores []struct {
		Id bson.ObjectId `bson:"_id" json:"id"`
		Score  int `json:"score"`
	}
	currentLatLng := [2]float64{lng, lat}
	c := context.DbCollection("groups")
	err := c.Find(bson.M{"loc":
	bson.M{"$geoWithin":
	bson.M{"$centerSphere": []interface{}{currentLatLng, 30 / 3963.2} }}}).Sort("-score").All(&groupScores);
	length := len(groupScores)

	for index,_ := range groupScores {
		groupScores[index].Score = length;
		length--;
	}
	return groupScores, err
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

	err = c.Insert(group)
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


func getCategoryName(id string, categories [] models.GroupCategory) string {
	for i := 0; i < len(categories); i++ {
		if(id == categories[i].Id.Hex()) {
			return categories[i].Name;
		}
	}
	return ""
}

func PopulateGroup() error {
	context := common.NewContext()
	defer context.Close()
	c := context.DbCollection("groups")

	var groups []models.Group
	err := c.Find(nil).All(&groups)

	if(err !=nil) {
		return err
	}

	categories, er := GetAllCategories();

	if(er != nil) {
		return err
	}

	groupsLen := len(groups)

	log.Print(groupsLen);

	for i := 0; i < groupsLen; i++ {
		groups[i].CategoryName = getCategoryName(groups[i].CategoryId.Hex(), categories);
		err = c.Update(bson.M{"_id": groups[i].Id},
			groups[i]);
		if(err != nil) {
			 log.Print(groups[i])
			 log.Print(err.Error())
		}
	}

	return err;
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

func GetAllGroups() (groups []models.Group, err error) {
	context := common.NewContext()
	defer context.Close()
	c := context.DbCollection("groups")

	err = c.Find(nil).All(&groups);
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



func RecalculateTrendingScore() (err error) {
	context := common.NewContext()
	defer context.Close()

	c := context.DbCollection("posts")



	if(err!= nil) {
		log.Print("Group Score Cron Error:", err.Error())
		return;
	}

	op := bson.M{
		"$group": bson.M{
			"_id": "$groupId",
			"sum_score": bson.M{
				"$sum": "$score",
			},
		},
	}

	operations := []bson.M{op}

	pipe := c.Pipe(operations);

	result := []bson.M{}
	err = pipe.All(&result)
	if err != nil {
		return err
	}

	cnt := common.NewContext()
	defer cnt.Close()

	groupsCol := cnt.DbCollection("groups")

	for _,res := range result {
		if res["_id"] != nil {
			_ = groupsCol.Update(bson.M{"_id": res["_id"]},
				bson.M{"$set": bson.M{
					"score": res["sum_score"],
				}})
		}
	}

	return
}
