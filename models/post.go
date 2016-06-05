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
	//UserId bson.ObjectId
	Id bson.ObjectId `bson:"_id" json:"id"`
	Loc        [2]float64  `bson:"loc" json:"loc"`
	Upvotes    int  `bson:"upvotes" json:"upvotes"`
	Downvotes  int  `bson:"downvotes" json:"downvotes"`
	CreatedOn  time.Time `bson:"createdOn" json:"-"`
	ModifiedOn time.Time `bson:"modifiedOn" json:"-"`
	GroupId    bson.ObjectId `bson:"groupId,omitempty" json:"groupId,omitempty"`
	GroupName   string `bson:"groupName,omitempty" json:"groupName,omitempty"`
	Content    string `bson:"content" json:"content"`
	ReplyCount int `bson:"replyCount" json:"replyCount"`
	ReplyIds   []bson.ObjectId `bson:"replyIds" json:"replyIds"`
	LabelId    bson.ObjectId `bson:"labelId,omitempty" json:"labelId,omitempty"`
	LabelName    string `bson:"labelName,omitempty" json:"labelName,omitempty"`
	isActive   bool `bson:"isActive" json:"isActive"`
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