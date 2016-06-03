package models

import (
	"time"
	//"github.com/mymachine8/fardo-api/bootstrap/dbconn"
	"gopkg.in/mgo.v2/bson"
	"io/ioutil"
	"path/filepath"
	"fmt"
	"os"
	"encoding/json"
)

type Post struct {
	//UserId bson.ObjectId
	Loc        [2]float64  `bson:"loc,omitempty" json:"loc,omitempty"`
	Upvotes    int  `bson:"upvotes,omitempty" json:"upvotes"`
	Downvotes  int  `bson:"downvotes,omitempty" json:"downvotes"`
	CreatedAt  time.Time `bson:"created_at,omitempty" json:"-"`
	ModifiedAt time.Time `json:"-"`
	GroupId    int `json:"groupId,omitempty"`
	CityId     int `json:"cityId,omitempty"`
	Content    string `json:"content,omitempty"`
	ReplyCount int `json:"replyCount,omitempty"`
	ReplyIds   []bson.ObjectId `json:"replyIds,omitempty"`
	LabelId    int `json:"labelId,omitempty"`
	isActive   bool
}

const PostCollectionName = "posts"

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