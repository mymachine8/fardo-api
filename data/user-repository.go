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
	}

	return user.Id.Hex(), err
}

func GetUserId(token string) (string, error) {
	context := common.NewContext()
	defer context.Close()
	c := context.DbCollection("access_tokens")

	var result struct{ userId bson.ObjectId `bson:"userId"` }

	err := c.Find(bson.M{"token": token}).Select(bson.M{"userID": 1}).One(result)

	return result.userId.Hex(), err

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

func UpdateUserGroup(token string, groupId string, lat float64, lng float64) ( bool, error) {
	isGroupLocked := true;
	context := common.NewContext()
	tokenCol := context.DbCollection("access_tokens")
	defer context.Close()

	userContext := common.NewContext()
	userCol := userContext.DbCollection("users")
	defer userContext.Close()

	groupContext := common.NewContext()
	groupCol := groupContext.DbCollection("groups")
	defer groupContext.Close()

	var result models.AccessToken
	err := tokenCol.Find(bson.M{"token": token}).One(&result)

	if (err != nil) {
		return isGroupLocked,err
	}

	var group models.Group

	err = groupCol.Find(bson.M{"loc":
	bson.M{"$geoWithin":
	bson.M{"$centerSphere": []interface{}{[2]float64{lng, lat}, 1 / 3963.2} }}}).One(&group);

	if(err == nil && (strings.Compare(group.Id.Hex(), groupId) == 0)) {
		isGroupLocked = false;
	}

	err = userCol.Update(bson.M{"_id": result.UserId},
		bson.M{"$set": bson.M{
			"groupId": bson.ObjectIdHex(groupId),
			"isGroupLocked" : isGroupLocked,
		}})

	if (err != nil) {
		return isGroupLocked,err
	}

	err = tokenCol.Update(bson.M{"userId": result.UserId},
		bson.M{"$set": bson.M{
			"groupId": bson.ObjectIdHex(groupId),
		}})

	return isGroupLocked,err
}