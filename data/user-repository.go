package data

import (
	"golang.org/x/crypto/bcrypt"
	"github.com/mymachine8/fardo-api/models"
	"gopkg.in/mgo.v2/bson"
	"github.com/mymachine8/fardo-api/common"
	"gopkg.in/mgo.v2"
	"log"
)

func RegisterUser(user models.User) error {
	context := common.NewContext()
	defer context.Close()
	c := context.DbCollection("users")
	obj_id := bson.NewObjectId()
	user.Id = obj_id
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

	var existingUser models.User;

	err = c.Find(bson.M{"imei": user.Imei}).One(&existingUser)

	if (err != nil && err.Error() == mgo.ErrNotFound.Error()) {
		err = c.Insert(&user)
		if (err != nil) {
			return
		}
		return user.Id.Hex(), err
	} else if(err!= nil){
          return
	}

	return existingUser.Id.Hex(), err
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
	accessToken.UserId = bson.ObjectIdHex(userId);
	err := c.Insert(&accessToken)

	return err
}

func UpdateUserGroup(token string, groupId string) (err error) {
	context := common.NewContext()
	tokenCol := context.DbCollection("access_tokens")
	defer context.Close()

	userContext := common.NewContext()
	userCol := userContext.DbCollection("users")
	defer userContext.Close()

	var result models.AccessToken
	err = tokenCol.Find(bson.M{"token": token}).One(&result)

	log.Print(result.UserId);
	if(err != nil) {
		return
	}

	err = userCol.Update(bson.M{"_id": result.UserId},
		bson.M{"$set": bson.M{
			"groupId": bson.ObjectIdHex(groupId),
		}})

	if(err != nil) {
		return
	}

	err = tokenCol.Update(bson.M{"userId": result.UserId},
		bson.M{"$set": bson.M{
			"groupId": bson.ObjectIdHex(groupId),
		}})

	return
}