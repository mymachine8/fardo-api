package models

import (
	"time"
	"gopkg.in/mgo.v2/bson"
)

const (
	NEGATIVE_VOTES_LIMIT = -4;
	SPAM_COUNT_LIMIT = 2;
)

type Post struct {
	Id                bson.ObjectId `bson:"_id" json:"id"`
	Loc               [2]float64  `bson:"loc" json:"loc"`
	City              string `bson:"city" json:"city"`
	State             string `bson:"state" json:"state"`
	Locality          string `bson:"locality" json:"locality"`
	FullAddress       string `bson:"fullAddress" json:"fullAddress"`
	ImageUrl          string `bson:"imageUrl" json:"imageUrl"`
	ImageType         string `bson:"imageType" json:"imageType"`
	ImageData         string `bson:"-" json:"imageData,omitempty"`
	ImageWidth        int `bson:"imageWidth,omitempty" json:"imageWidth,omitempty"`
	ImageHeight       int `bson:"imageHeight,omitempty" json:"imageHeight,omitempty"`
	Upvotes           int  `bson:"upvotes" json:"upvotes"`
	Downvotes         int  `bson:"downvotes" json:"downvotes"`
	Votes             int  `bson:"votes" json:"votes"`
	SpamCount         int  `bson:"spamCount" json:"-"`
	SpamReasons       []string `bson:"spamReasons" json:"-"`
	Score             float64  `bson:"score" json:"score"`
	CreatedOn         time.Time `bson:"createdOn" json:"createdOn"`
	UserId            bson.ObjectId `bson:"userId" json:"userId"`
	Username          string `bson:"username" json:"username"`
	Jid               string `bson:"jid,omitempty" json:"jid,omitempty"`
	VoteClicked       string `bson:"-" json:"voteClicked,omitempty"`
	IsAnonymous       bool `bson:"isAnonymous" json:"isAnonymous"`
	ModifiedOn        time.Time `bson:"modifiedOn,omitempty" json:"-"`
	GroupId           bson.ObjectId `bson:"groupId,omitempty" json:"groupId,omitempty"`
	GroupName         string `bson:"groupName,omitempty" json:"groupName,omitempty"`
	GroupCategoryName string `bson:"groupCategoryName,omitempty" json:"-"`
	PlaceName         string `bson:"placeName,omitempty" json:"placeName,omitempty"`
	PlaceType         string `bson:"placeType,omitempty" json:"placeType,omitempty"`
	Content           string `bson:"content" json:"content"`
	ReplyCount        int `bson:"replyCount" json:"replyCount"`
	IsGroup           bool `bson:"isGroup" json:"isGroup"`
	IsLocation        bool `bson:"isLocation" json:"isLocation"`
	IsActive          bool `bson:"isActive" json:"isActive"`
}

type ScoreSorter []Post

func (a ScoreSorter) Len() int {
	return len(a)
}
func (a ScoreSorter) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}
func (a ScoreSorter) Less(i, j int) bool {
	return a[i].Score > a[j].Score
}