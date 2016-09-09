package data

import (
	"time"
	"github.com/mymachine8/fardo-api/models"
	"gopkg.in/mgo.v2/bson"
	"github.com/mymachine8/fardo-api/common"
	"strings"
	"encoding/base64"
	"math"
)

func GetPopularGroups(lat float64, lng float64) (groups []models.GroupLite, err error) {
	//TODO: Get based on number of posts in the last 2 days
	//TODO: Maintain Global popular groups in a buffer

	var localGroups []models.GroupLite;
	var globalGroups []models.GroupLite;

	localGroups, err = GetNearByPopularGroups(lat, lng);
	globalGroups, err = GetGlobalPopularGroups();

	var i int;
	for i = 0; i < 2 && i < len(localGroups); i++ {
		groups = append(groups, localGroups[i]);
	}

	remaining := 2 - i;

	for i = 0; i < 3 + remaining && i < len(globalGroups); i++ {
		if (!idInGroups(globalGroups[i].Id.Hex(), groups)) {
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
	err = c.Find(nil).Sort("-score").Select(bson.M{"name": 1, "shortName": 1, "categoryName" : 1}).Limit(5).All(&groups)
	return;
}

func GetNearByPopularGroups(lat float64, lng float64) (groups []models.GroupLite, err error) {
	context := common.NewContext()
	defer context.Close()

	currentLatLng := [2]float64{lng, lat}
	c := context.DbCollection("groups")
	err = c.Find(bson.M{"loc":
	bson.M{"$geoWithin":
	bson.M{"$centerSphere": []interface{}{currentLatLng, 40 / 3963.2} }}}).Select(bson.M{"name": 1, "shortName": 1, "categoryName" : 1}).Limit(5).All(&groups);
	return
}

func GetNearByGroups(lat float64, lng float64,limit int64) (groups []models.Group, err error) {
	context := common.NewContext()
	defer context.Close()

	currentLatLng := [2]float64{lng, lat}
	c := context.DbCollection("groups")
	query := c.Find(bson.M{"loc":
	bson.M{"$geoWithin":
	bson.M{"$centerSphere": []interface{}{currentLatLng, 40 / 3963.2} }}});
	if(limit > 0) {
		err = query.Limit(int(limit)).All(&groups);
	} else {
		err = query.All(&groups);
	}
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

	if (err != nil) {
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

func GetGroups(page int, groupParams models.Group) (groups []models.Group, err error) {
	context := common.NewContext()
	defer context.Close()
	c := context.DbCollection("groups")

	skip := page * 20;
	params := make(map[string]interface{})
	if (len(groupParams.Name) > 0) {
		params["name"] = bson.RegEx{Pattern: groupParams.Name, Options: "i"};
	}
	if (len(groupParams.City) > 0) {
		params["city"] = bson.RegEx{Pattern: groupParams.City, Options: "i"};
	}
	if (len(groupParams.State) > 0) {
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

func CalculatePlacesTrendingScore() (err error) {
	context := common.NewContext()
	defer context.Close()
	c := context.DbCollection("groups")

	postsContext := common.NewContext()
	defer postsContext.Close()
	postsCol := postsContext.DbCollection("posts")

	pipe := c.Pipe([]bson.M{{"$project": bson.M{"scoreLastUpdated": 1, "scores": 1, "scoreCount": bson.M{"$size": "$scores" }}}})

	var result []struct {
		Id               bson.ObjectId `bson:"_id" json:"id"`
		ScoreCount       int `bson:"scoreCount" json:"scoreCount"`
		Scores           []int `bson:"scores" json:"_"`
		ScoreLastUpdated time.Time `bson:"scoreLastUpdated" json:"scoreLastUpdated,omitempty"`
	}

	err = pipe.All(&result);

	now := time.Now().UTC()

	then := now.AddDate(0, 0, -1)

	pipe = postsCol.Pipe([]bson.M{{"$match": bson.M{"createdOn":bson.M{"$gt" : then, "$lt": now}}}, {"$group": bson.M{"_id": "$groupId", "count": bson.M{"$sum": 1 }}}})

	var postsResult []struct {
		Id    bson.ObjectId `bson:"_id" json:"id"`
		Count int `bson:"count" json:"count"`
	}

	err = pipe.All(&postsResult);

	for _, group := range result {
		count := 0
		for _, grp := range postsResult {
			if (grp.Id.Hex() == group.Id.Hex()) {
				count = grp.Count
			}
		}
		trendingScore := calculateZIndex(count, group.Scores)

		if (now.Day() != group.ScoreLastUpdated.Day()) {
			c.Update(bson.M{"_id": group.Id},
				bson.M{"$pop": bson.M{"scores": -1 } });
		} else {
			c.Update(bson.M{"_id": group.Id},
				bson.M{"$pop": bson.M{"scores": 1 } });
		}
		c.Update(bson.M{"_id": group.Id},
			bson.M{"$push": bson.M{"scores": count },
				"$set": bson.M{"scoreLastUpdated" : now, "trendingScore" : trendingScore }});
	}

	return err
}

func calculateZIndex(currentScore int, prevScores [] int) float64 {
	n := len(prevScores)

	var sum float64

	sum = 0

	for i := 0; i < n; i++ {
		sum = sum + float64(prevScores[i])
	}

	var avg float64
	avg = float64(sum / float64(n))

	if (avg == 0) {
		return -999999
	}

	sum = 0

	for i := 0; i < n; i++ {
		sum += math.Pow((float64(prevScores[i]) - avg), 2)
	}

	variance := sum / float64(n)

	standardDeviation := math.Sqrt(variance)

	zScore := (float64(currentScore) - avg) / standardDeviation

	return zScore
}
