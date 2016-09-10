package models

import (
	"gopkg.in/mgo.v2/bson"
	"time"
)

type StatusType string

const (
	AppNotAvailable StatusType = "AppNotAvailable"
	Suspended StatusType = "Suspended"
	Inactive StatusType = "Inactive"
	Active StatusType = "Active"
)

type User struct {
	Id                bson.ObjectId `json:"id,omitempty" bson:"_id,omitempty"`
	Username          string `bson:"username,omitempty" json:"username,omitempty"`
	SessionId         uint64  `bson:"sessionId,omitempty" json:"sessionId,omitempty"`
	Imei              string `json:"imei,omitempty" bson:"imei,omitempty"`
	Status            string `json:"status" bson:"status"`
	FcmToken          string `bson:"fcmToken" json:"fcmToken"`
	Token             string `bson:"token" json:"token"`
	TokenSecret       string `bson:"tokenSecret" json:"tokenSecret"`
	Score             int `json:"score" bson:"score"`
	SpamPostCount     int `json:"spamPostCount" bson:"spamPostCount"`
	DownvotePostCount int `json:"downvotePostCount" bson:"downvotePostCount"`
	Loc               [2]float64 `bson:"loc" json:"loc"`
	HomeLoc           [2]float64 `bson:"homeLoc" json:"homeLoc"`
	HomeAddress       string `json:"homeAddress,omitempty" bson:"homeAddress,omitempty"`
	Phone             string `json:"phone,omitempty" bson:"phone,omitempty"`
	GroupId           bson.ObjectId `json:"groupId,omitempty" bson:"groupId,omitempty"`
	IsGroupLocked     bool `json:"isGroupLocked,omitempty" bson:"isGroupLocked,omitempty"`
	IsActive          bool `bson:"isActive" json:"isActive"`
	Role              string `bson:"role,omitempty"`
	CreatedOn         time.Time `bson:"createdOn" json:"-"`
	ModifiedOn        time.Time `bson:"modifiedOn" json:"-"`
}
