package data

import (
	"github.com/mymachine8/fardo-api/models"
	"gopkg.in/mgo.v2/bson"
	"github.com/mymachine8/fardo-api/common"
	"log"
	"time"
	"gopkg.in/mgo.v2"
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

func RegisterAppUser(user models.User) (models.User, error) {

	context := common.NewContext()
	defer context.Close()
	c := context.DbCollection("users")

	user.IsActive = true;
	user.ModifiedOn = time.Now().UTC()
	user.CreatedOn = time.Now().UTC()
	user.Token = bson.NewObjectId().Hex();
	user.Score = 200;
	user.CreatedOn = time.Now().UTC()
	user.Id = bson.NewObjectId()
	err := c.Update(bson.M{"imei": user.Imei, "phone": bson.M{"$exists" : false}},
		bson.M{"$set": bson.M{
			"isActive": false,
			"modifiedOn": time.Now().UTC(),
		}})

	err = c.Insert(&user)

	return user, err
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

func SetUsernameToken(accessToken string, username string) (string, error) {

	userContext := common.NewContext()
	userCol := userContext.DbCollection("users")
	defer userContext.Close()

	var users []models.User
	err := userCol.Find(bson.M{"username": username}).All(&users)

	if (err != nil) {
		return "", err
	}

	if (len(users) > 0) {
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

	if (err != nil) {
		return false, err
	}

	if (len(users) > 0) {
		return false, err
	}

	return true, err
}

func ChangeUserPhone (accessToken string, imei string, sessionId uint64, token string, tokenSecret string, phone string, fcmToken string) (models.User, models.Group, error) {
	userContext := common.NewContext()
	userCol := userContext.DbCollection("users")
	defer userContext.Close()

	var existing models.User
	err := userCol.Find(bson.M{"phone": phone, "isActive": true}).One(&existing)
	if (err != nil && err.Error() == mgo.ErrNotFound.Error()) {
		err = userCol.Update(bson.M{"token": accessToken},
			bson.M{"$set": bson.M{
				"token": token,
				"imei": imei,
				"tokenSecret": tokenSecret,
				"phone" : phone,
				"sessionId" : sessionId,
				"fcmToken" : fcmToken,
			}})
	} else {
		err = userCol.Update(bson.M{"phone": phone, "isActive" : true},
			bson.M{"$set": bson.M{
				"token": token,
				"imei": imei,
				"tokenSecret": tokenSecret,
				"phone" : phone,
				"sessionId" : sessionId,
				"fcmToken" : fcmToken,
			}})
		err = userCol.Update(bson.M{"imei": imei, "phone": bson.M{"$exists" : false}},
			bson.M{"$set": bson.M{
				"isActive": false,
				"modifiedOn": time.Now().UTC(),
			}})
	}

	var user models.User

	err = userCol.Find(bson.M{"token": token}).One(&user)
	var group models.Group
	if (err == nil) {
		if (len(user.GroupId) > 0) {
			group, err = GetGroupById(user.GroupId.Hex())
		}
	}

	return user, group, err
}

func ChangeUserHomeLocation(accessToken string, homeAddress string, lat float64, lng float64) error {
	userContext := common.NewContext()
	userCol := userContext.DbCollection("users")
	defer userContext.Close()

	LatLng := [2]float64{lng, lat}
	err := userCol.Update(bson.M{"token": accessToken},
		bson.M{"$set": bson.M{
			"homeLoc": LatLng,
			"homeAddress" : homeAddress,
		}})

	return err
}

func GetUserScore(accessToken string) (score int, err error) {
	userContext := common.NewContext()
	userCol := userContext.DbCollection("users")
	defer userContext.Close()

	var user models.User;
	err = userCol.Find(bson.M{"token": accessToken}).One(&user)

	if (err != nil) {
		return score, err;
	}

	return user.Score, err
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

func isLocationInGroup(groupId string, lat float64, lng float64) (isNear bool) {
	groupContext := common.NewContext()
	groupCol := groupContext.DbCollection("groups")
	defer groupContext.Close()
	var group models.Group
	err := groupCol.FindId(bson.ObjectIdHex(groupId)).One(&group);

	log.Print(common.DistanceLatLong(group.Loc[1], lat, group.Loc[0], lng));

	if (err == nil && common.DistanceLatLong(group.Loc[1], lat, group.Loc[0], lng) < float64(group.Radius)) {
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

	if (err == nil && !isLock) {
		var user models.User
		_ = userCol.Find(bson.M{"token": token}).One(&user)
		common.GroupUnlockedNotification(user)
	}

	return err;
}

func UpdateUserGroup(token string, groupId string, lat float64, lng float64) (bool, error) {

	userContext := common.NewContext()
	userCol := userContext.DbCollection("users")
	defer userContext.Close()

	isNear := isLocationInGroup(groupId, lat, lng);

	err := userCol.Update(bson.M{"token": token},
		bson.M{"$set": bson.M{
			"groupId": bson.ObjectIdHex(groupId),
			"isGroupLocked" : !isNear,
		}})

	return !isNear, err
}

func RemoveUserGroup(token string) (error) {

	userContext := common.NewContext()
	userCol := userContext.DbCollection("users")
	defer userContext.Close()

	err := userCol.Update(bson.M{"token": token},
		bson.M{"$unset": bson.M{
			"groupId": 1,
		}})

	return err
}

func RemoveHomeLocation(token string) (error) {

	userContext := common.NewContext()
	userCol := userContext.DbCollection("users")
	defer userContext.Close()

	err := userCol.Update(bson.M{"token": token},
		bson.M{"$unset": bson.M{
			"homeLoc": 1,
			"homeAddress" : 1,
		}})

	return err
}

func GetUserInfo(token string) (user models.User, err error) {

	userContext := common.NewContext()
	userCol := userContext.DbCollection("users")
	defer userContext.Close()

	err = userCol.Find(bson.M{"token": token}).One(&user)

	return

}

func GetUsers() (users [] models.User, err error) {
	userContext := common.NewContext()
	userCol := userContext.DbCollection("users")
	defer userContext.Close()

	err = userCol.Find(bson.M{"isActive": true}).All(&users)

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

func SetUserFeedback(token string, content string, phone string, email string) error {
	context := common.NewContext()
	defer context.Close()
	c := context.DbCollection("user_feedback")

	userContext := common.NewContext()
	userCol := userContext.DbCollection("users")
	defer userContext.Close()

	var user models.User;

	err := userCol.Find(bson.M{"token": token}).One(&user)

	if (err != nil) {
		return err;
	}

	params := bson.M{};

	params["_id"] = bson.NewObjectId()
	params["userId"] = user.Id;
	params["content"] = content;
	params["createdOn"] = time.Now().UTC();

	if (len(phone) > 0 ) {
		params["phone"] = phone;
	}

	if (len(email) > 0) {
		params["email"] = email;
	}

	err = c.Insert(params)

	if (err != nil) {
		log.Print(err.Error())
	}

	common.FeedbackNotification(user.FcmToken)

	return err
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