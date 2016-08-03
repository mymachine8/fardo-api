package common

import (
	"github.com/NaySoftware/go-fcm"
	"fmt"
	"github.com/mymachine8/fardo-api/models"
	"gopkg.in/mgo.v2/bson"
	"log"
)

const (
	serverKey = "AIzaSyDcXHtO1YB-tCRwFAvaVqJDud8gR00VJs4"
)

func SendUpvoteNotification(post models.Post) {
	token, err := findUserToken(post.UserId.Hex());
	if (err != nil) {
		log.Print(err.Error())
	}

	data := map[string]string{
		"id": bson.NewObjectId().Hex(),
		"postId": post.Id.Hex(),
		"time": post.ModifiedOn.String(),
	}

	log.Print(data)

	message := "You got " + string(post.Upvotes) + " for your post " + string(post.Content)

	log.Print(message)

	ids := []string{token.FcmToken}

	log.Print(ids)

	sendNotification(ids, message, data);
}

func SendCommentUpvoteNotification(comment models.Comment) {
	token, err := findUserToken(comment.UserId.Hex());
	if (err != nil) {
		return;
	}

	data := map[string]string{
		"id": bson.NewObjectId().Hex(),
		"postId": comment.PostId.Hex(),
		"commentId": comment.Id.Hex(),
		"time": comment.ModifiedOn.String(),
	}

	message := "You got " + string(comment.Upvotes) + " for your comment " + string(comment.Content)

	ids := []string{token.FcmToken}

	sendNotification(ids, message, data);
}

func findUserById(userId string) (user models.User, err error) {
	context := NewContext()
	defer context.Close()
	c := context.DbCollection("users")

	err = c.FindId(bson.ObjectIdHex(userId)).One(&user);
	return
}

func findUserToken(userId string)(token models.AccessToken, err error) {
	context := NewContext()
	defer context.Close()
	c := context.DbCollection("access_tokens")

	err = c.Find(bson.M{"userId": bson.ObjectIdHex(userId)}).One(&token);
	return
}

func findNearByUsers(lat float64, lng float64) (users []models.User, err error) {
	context := NewContext()
	defer context.Close()
	c := context.DbCollection("users")

	currentLatLng := [2]float64{lng, lat}
	err = c.Find(bson.M{"loc":
	bson.M{"$geoWithin":
	bson.M{"$centerSphere": []interface{}{currentLatLng, 0.05 / 3963.2} }}}).All(&users);
	return
}

func SendCommentNotification(post models.Post, comment models.Comment) {
	user, err := findUserById(post.UserId.Hex());
	if (err != nil) {
		return;
	}

	data := map[string]string{
		"id": bson.NewObjectId().Hex(),
		"postId": post.Id.Hex(),
		"commentId": comment.Id.Hex(),
		"time": comment.CreatedOn.String(),
	}

	message := "Someone commented on your post " + post.Content;

	ids := []string{user.Id.Hex()}

	sendNotification(ids, message, data);

}

func SendReplyNotification(comment models.Comment, reply models.Reply) {
	user, err := findUserById(comment.UserId.Hex());
	if (err != nil) {
		return;
	}

	data := map[string]string{
		"id": bson.NewObjectId().Hex(),
		"postId": comment.PostId.Hex(),
		"commentId": comment.Id.Hex(),
		"time": comment.CreatedOn.String(),
	}

	message := "Someone replied to your comment " + comment.Content;

	ids := []string{user.Id.Hex()}

	sendNotification(ids, message, data);
}

func SendNearByNotification(post models.Post) {
	users, err := findNearByUsers(post.Loc[1], post.Loc[0]);

	if (err != nil) {
		return;
	}

	ids := []string{}

	for _, user := range users {
		ids = append(ids, user.FcmToken);
	}

	data := map[string]string{
		"id": bson.NewObjectId().Hex(),
		"postId": post.Id.Hex(),
		"time": post.CreatedOn.String(),
	}

	message := "Someone nearby posted " + post.Content;

	sendNotification(ids, message, data);
}

func SendDeletePostNotification(post models.Post) {
	token, err := findUserToken(post.UserId.Hex());
	if (err != nil) {
		return;
	}

	data := map[string]string{
		"id": bson.NewObjectId().Hex(),
		"postId": post.Id.Hex(),
		"time": post.ModifiedOn.String(),
	}

	message :=  "Your Post " + post.Content + " has been suspended as people in your community downvoted or reported about it"
	ids := []string{token.FcmToken}

	sendNotification(ids, message, data);
}

func AppLocationReadyNotification() {
	//TODO: Send Ready Notification, when the app is ready
	//message := "This App is now available in your location!!!"
}

func sendNotification(fcmTokens []string, message string, data map[string]string) {

	c := fcm.NewFcmClient(serverKey)
	log.Print("fcmTokens")
	log.Print(fcmTokens)
	c.NewFcmRegIdsMsg(fcmTokens, data)
	var notification fcm.NotificationPayload;
	notification.Title = "Zing";
	notification.Body = message;
	c.SetNotificationPayload(&notification);

	status, err := c.Send()

	if err == nil {
		status.PrintResults()
	} else {
		fmt.Println("FCM Error: " + err.Error());
	}
}
