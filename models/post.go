package models

import (
	"time"
	"gopkg.in/mgo.v2/bson"
	"io/ioutil"
	"path/filepath"
	"fmt"
	"os"
	"encoding/json"
)

type Post struct {
	Id         bson.ObjectId `bson:"_id" json:"id"`
	Loc        [2]float64  `bson:"loc" json:"loc"`
	City string `bson:"city" json:"city"`
	State string `bson:"state" json:"state"`
	Locality string `bson:"locality" json:"locality"`
	FullAddress string `bson:"fullAddress" json:"fullAddress"`
	ImageUrl string `bson:"imageUrl" json:"imageUrl"`
	ImageType string `bson:"imageType" json:"imageType"`
	ImageData string `bson:"-" json:"imageData, omitempty"`
	Upvotes    int  `bson:"upvotes" json:"upvotes"`
	Downvotes  int  `bson:"downvotes" json:"downvotes"`
	CreatedOn  time.Time `bson:"createdOn" json:"createdOn"`
	UserId     bson.ObjectId `bson:"userId" json:"-"`
	ModifiedOn time.Time `bson:"modifiedOn" json:"-"`
	GroupId    bson.ObjectId `bson:"groupId,omitempty" json:"groupId,omitempty"`
	GroupName  string `bson:"groupName,omitempty" json:"groupName,omitempty"`
	Content    string `bson:"content" json:"content"`
	ReplyCount int `bson:"replyCount" json:"replyCount"`
	LabelId    bson.ObjectId `bson:"labelId,omitempty" json:"labelId,omitempty"`
	LabelName  string `bson:"labelName,omitempty" json:"labelName,omitempty"`
	IsGroup bool `bson:"isGroup" json:"isGroup"`
	IsLocation bool `bson:"isLocation" json:"isLocation"`
	IsActive   bool `bson:"isActive" json:"isActive"`
}

func MyPosts(userId bson.ObjectId, currentLatLng [2]float64) []Post {
	//TODO: Get the feed from his Groups, 1 km of his current location.

	/*mongoSession := dbconn.GetInstance();
	s := mongoSession.GetSession();
	c := s.DB(mongoSession.GetDatabaseName()).C(PostCollectionName);
	result := Post {}
	err := c.Find(bson.M{"loc":
			bson.M{"$geoWithin":
			bson.M{"$center": []interface{}{currentLatLng, 1} }}}).All(&result);
	if err != nil {
		panic(err)
	}*/
	absPath, _ := filepath.Abs("./models/sample-posts.json")
	file, e := ioutil.ReadFile(absPath)
	if e != nil {
		fmt.Printf("File error: %v\n", e)
		os.Exit(1)
	}
	fmt.Printf("%s\n", string(file))

	var posts []Post;

	json.Unmarshal(file, &posts);

	return posts;
}