package data

import (
	"github.com/mymachine8/fardo-api/models"
	"gopkg.in/mgo.v2/bson"
	"github.com/mymachine8/fardo-api/common"
	"log"
	"time"
	"strings"
)

type ActionType int8

const (
	ActionCreate ActionType = 0
	ActionDownvote ActionType = 1
	ActionSpam ActionType = 2
	ActionShare ActionType = 3
	ActionUpvote ActionType = 4
)

func RegisterUser(user models.User) error {
	context := common.NewContext()
	defer context.Close()
	c := context.DbCollection("users")
	user.Id = bson.NewObjectId()
	user.CreatedOn = time.Now().UTC()
	err := c.Insert(&user)
	return err
}

func Login(user models.User) (u models.User, err error) {
	context := common.NewContext()
	defer context.Close()
	c := context.DbCollection("users")

	err = c.Find(bson.M{"username": user.Username}).One(&u)
	if err != nil {
		u = models.User{}
	}
	return
}

func RegisterAppUser(user models.User) (models.User,  error) {

	context := common.NewContext()
	defer context.Close()
	c := context.DbCollection("users")


	user.IsActive = true;
	user.CreatedOn = time.Now().UTC()
	user.ModifiedOn = time.Now().UTC()
	user.Token = bson.NewObjectId().Hex();

	changeInfo, err := c.Upsert(bson.M{"imei": user.Imei}, user);

	if(err != nil) {
		return user, err;
	}

	id, _ := changeInfo.UpsertedId.(bson.ObjectId)

	user.Id = id;

	return user, err;
}

func SetUserToken(token string, username string) error {
	context := common.NewContext()
	defer context.Close()
	c := context.DbCollection("users")

	err := c.Update(bson.M{"username": username},
		bson.M{"$set": bson.M{
			"token": token,
			"modifiedOn": time.Now().UTC(),
		}})

	return err
}

func SetUserFcmToken(accessToken string, fcmToken string) error {

	userContext := common.NewContext()
	userCol := userContext.DbCollection("users")
	defer userContext.Close()

	err := userCol.Update(bson.M{"token": accessToken},
		bson.M{"$set": bson.M{
			"fcmToken": fcmToken,
		}})

	return err
}


func SetUsernameToken(accessToken string , username string) (string, error) {

	userContext := common.NewContext()
	userCol := userContext.DbCollection("users")
	defer userContext.Close()

	var users []models.User
	err := userCol.Find(bson.M{"username": username}).All(&users)

	if(err != nil) {
		return "", err
	}

	if(len(users) > 0) {
		return "username already exists", err
	}

	err = userCol.Update(bson.M{"token": accessToken},
		bson.M{"$set": bson.M{
			"username": username,
		}})

	return "success", err
}

func CheckUsernameAvailability(username string) (bool, error) {

	userContext := common.NewContext()
	userCol := userContext.DbCollection("users")
	defer userContext.Close()

	var users []models.User
	err := userCol.Find(bson.M{"username": username}).All(&users)

	if(err != nil) {
		return false, err
	}

	if(len(users) > 0) {
		return false, err
	}

	return true, err
}

func ChangeUserPhone(accessToken string, sessionId uint64, token string, tokenSecret string, phone string) error {
	userContext := common.NewContext()
	userCol := userContext.DbCollection("users")
	defer userContext.Close()

	err := userCol.Update(bson.M{"token": accessToken},
		bson.M{"$set": bson.M{
			"token": token,
			"tokenSecret": tokenSecret,
			"phone" : phone,
			"sessionId" : sessionId,
		}})

	return err
}

func SetUserLocation(accessToken string, lat float64, lng float64) error {

	userContext := common.NewContext()
	userCol := userContext.DbCollection("users")
	defer userContext.Close()

	LatLng := [2]float64{lng, lat}
	err := userCol.Update(bson.M{"token": accessToken},
		bson.M{"$set": bson.M{
			"loc": LatLng,
		}})

	return err
}

func isLocationInGroup(groupId string, lat float64, lng float64)(isNear bool) {
	groupContext := common.NewContext()
	groupCol := groupContext.DbCollection("groups")
	defer groupContext.Close()
	var group models.Group
	err := groupCol.Find(bson.M{"loc":
	bson.M{"$geoWithin":
	bson.M{"$centerSphere": []interface{}{[2]float64{lng, lat}, 1 / 3963.2} }}}).One(&group);

	if (err == nil && (strings.Compare(group.Id.Hex(), groupId) == 0)) {
		return true
	}

	return false
}

func LockUserGroup(token string, isLock bool) error {

	userContext := common.NewContext()
	userCol := userContext.DbCollection("users")
	defer userContext.Close()

	err := userCol.Update(bson.M{"token": token},
		bson.M{"$set": bson.M{
			"isGroupLocked" : isLock,
		}})

	return err;
}

func UpdateUserGroup(token string, groupId string, lat float64, lng float64) (bool, error) {
	isGroupLocked := true;

	userContext := common.NewContext()
	userCol := userContext.DbCollection("users")
	defer userContext.Close()

	isNear := isLocationInGroup(groupId,lat,lng);

	err := userCol.Update(bson.M{"token": token},
		bson.M{"$set": bson.M{
			"groupId": bson.ObjectIdHex(groupId),
			"isGroupLocked" : !isNear,
		}})

	return isGroupLocked, err
}

func GetUserInfo(token string) (user models.User, err error){

	userContext := common.NewContext()
	userCol := userContext.DbCollection("users")
	defer userContext.Close()

	err = userCol.Find(bson.M{"token": token}).One(&user)

	return

}

func CalculateUserScore(post models.Post, actionType ActionType) {
	var resultScore int;
	switch actionType {
	case ActionCreate:
		resultScore = postCreateScore();
		break;
	case ActionDownvote:
		resultScore = downvoteScore(post.Upvotes - post.Downvotes) * -1;
		break;
	case ActionSpam:
		resultScore = spamScore(post.SpamCount) * -1;
		break;
	case ActionShare:
		resultScore = shareScore();
		break;
	case ActionUpvote:
		resultScore = upvoteScore();
		break;
	}

	context := common.NewContext()
	defer context.Close()
	c := context.DbCollection("users")

	err := c.Update(bson.M{"_id": post.UserId},
		bson.M{"$inc": bson.M{
			"score": resultScore,
		}})

	if (err != nil) {
		log.Print(err.Error())
	}
}

func postCreateScore() int {
	return 10;
}

func downvoteScore(votes int) int {
	if (votes == 1 || votes == 2) {
		return 5;
	}
	if (votes == 3) {
		return 20;
	}
	return 0;
}

func spamScore(spamCount int) int {
	return 10 * spamCount;
}

func shareScore() int {
	return 10;
}

func upvoteScore() int {
	return 5;
}