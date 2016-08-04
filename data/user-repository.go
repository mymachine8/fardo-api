package data

import (
	"golang.org/x/crypto/bcrypt"
	"github.com/mymachine8/fardo-api/models"
	"gopkg.in/mgo.v2/bson"
	"github.com/mymachine8/fardo-api/common"
	"gopkg.in/mgo.v2"
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
	hpass, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}
	user.HashPassword = hpass
	//clear the incoming text password
	user.Password = ""
	err = c.Insert(&user)
	return err
}

func Login(user models.User) (u models.User, err error) {
	context := common.NewContext()
	defer context.Close()
	c := context.DbCollection("users")

	err = c.Find(bson.M{"username": user.Username}).One(&u)
	if err != nil {
		return
	}

	// Validate password
	err = bcrypt.CompareHashAndPassword(u.HashPassword, []byte(user.Password))
	if err != nil {
		u = models.User{}
	}
	return
}

func RegisterAppUser(user models.User) (userId string, err error) {

	lastLocation := user.Loc;
	context := common.NewContext()
	defer context.Close()
	c := context.DbCollection("users")

	err = c.Find(bson.M{"imei": user.Imei}).One(&user)

	if (err != nil && err.Error() == mgo.ErrNotFound.Error()) {
		user.Id = bson.NewObjectId()
		user.CreatedOn = time.Now().UTC()
		err = c.Insert(&user)
		if (err != nil) {
			return
		}
		return user.Id.Hex(), err
	} else if (err != nil) {
		return
	} else {
		err = c.Update(bson.M{"_id": user.Id},
			bson.M{"$set": bson.M{
				"lastKnownLocation": lastLocation,
			}})
	}

	return user.Id.Hex(), err
}

func GetAccessTokenDetails(token string) (result models.AccessToken, err error) {
	context := common.NewContext()
	defer context.Close()
	c := context.DbCollection("access_tokens")

	err = c.Find(bson.M{"token": token}).One(&result)

	return

}

func SetUserToken(token string, userId string) error {
	log.Println(userId);
	context := common.NewContext()
	defer context.Close()
	c := context.DbCollection("access_tokens")

	var accessToken models.AccessToken;
	accessToken.Id = bson.NewObjectId();
	accessToken.Token = token;
	accessToken.CreatedOn = time.Now().UTC()
	accessToken.UserId = bson.ObjectIdHex(userId);
	_, err := c.Upsert(bson.M{"userId": accessToken.UserId},
		bson.M{"$set": bson.M{
			"token": token,
			"userId": bson.ObjectIdHex(userId),
			"createdOn": accessToken.CreatedOn,
		}})

	return err
}

func SetUserFcmToken(accessToken string, fcmToken string) error {
	context := common.NewContext()
	defer context.Close()
	c := context.DbCollection("access_tokens")

	err := c.Update(bson.M{"token": accessToken},
		bson.M{"$set": bson.M{
			"fcmToken": fcmToken,
		}})

	if (err != nil) {
		log.Print(err.Error())
		return err
	}

	var result models.AccessToken
	err = c.Find(bson.M{"token": accessToken}).One(&result)

	userContext := common.NewContext()
	userCol := userContext.DbCollection("users")
	defer userContext.Close()

	err = userCol.Update(bson.M{"_id": result.UserId},
		bson.M{"$set": bson.M{
			"fcmToken": fcmToken,
		}})

	return err
}

func SetUserLocation(accessToken string, lat float64, lng float64) error {
	context := common.NewContext()
	defer context.Close()
	c := context.DbCollection("access_tokens")
	var result models.AccessToken
	err := c.Find(bson.M{"token": accessToken}).One(&result)
	if (err != nil) {
		return err;
	}
	userContext := common.NewContext()
	userCol := userContext.DbCollection("users")
	defer userContext.Close()

	LatLng := [2]float64{lng, lat}
	err = userCol.Update(bson.M{"_id": result.UserId},
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
	context := common.NewContext()
	tokenCol := context.DbCollection("access_tokens")
	defer context.Close()

	userContext := common.NewContext()
	userCol := userContext.DbCollection("users")
	defer userContext.Close()


	var result models.AccessToken
	err := tokenCol.Find(bson.M{"token": token}).One(&result)

	if(err !=nil) {
		return err
	}

	err = userCol.Update(bson.M{"_id": result.UserId},
		bson.M{"$set": bson.M{
			"isGroupLocked" : isLock,
		}})

	return err;
}

func UpdateUserGroup(token string, groupId string, lat float64, lng float64) (bool, error) {
	isGroupLocked := true;
	context := common.NewContext()
	tokenCol := context.DbCollection("access_tokens")
	defer context.Close()

	userContext := common.NewContext()
	userCol := userContext.DbCollection("users")
	defer userContext.Close()


	var result models.AccessToken
	err := tokenCol.Find(bson.M{"token": token}).One(&result)

	if (err != nil) {
		return isGroupLocked, err
	}

	isNear := isLocationInGroup(groupId,lat,lng);

	err = userCol.Update(bson.M{"_id": result.UserId},
		bson.M{"$set": bson.M{
			"groupId": bson.ObjectIdHex(groupId),
			"isGroupLocked" : !isNear,
		}})

	if (err != nil) {
		return isGroupLocked, err
	}

	err = tokenCol.Update(bson.M{"userId": result.UserId},
		bson.M{"$set": bson.M{
			"groupId": bson.ObjectIdHex(groupId),
		}})

	return isGroupLocked, err
}

func GetUserInfo(token string) (user models.User, err error){
	context := common.NewContext()
	tokenCol := context.DbCollection("access_tokens")
	defer context.Close()

	userContext := common.NewContext()
	userCol := userContext.DbCollection("users")
	defer userContext.Close()

	var result models.AccessToken
	err = tokenCol.Find(bson.M{"token": token}).One(&result)

	if (err != nil) {
		return
	}

	err = userCol.Find(bson.M{"_id": result.UserId}).One(&user)

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