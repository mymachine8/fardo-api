package common

import (
	"fmt"
	"github.com/mymachine8/fardo-api/models"
	"gopkg.in/mgo.v2/bson"
	"log"
	"strconv"
	"time"
	"net/http"
	"bytes"
	"io/ioutil"
	"encoding/json"
	"math/rand"
	"github.com/mymachine8/fardo-api/slack"
)

const (
	serverKey = "AIzaSyCwq8rOk4n96o3cS1mnFL1bGcfky4CL9es"
)

const (
	// fcm_server_url fcm server url
	fcm_server_url = "https://fcm.googleapis.com/fcm/send"
	// MAX_TTL the default ttl for a notification
	MAX_TTL = 2419200
	// Priority_HIGH notification priority
	Priority_HIGH = "high"
	// Priority_NORMAL notification priority
	Priority_NORMAL = "normal"
	// retry_after_header header name
	retry_after_header = "Retry-After"
	// error_key readable error caching !
	error_key = "error"
)

var (
	// retreyableErrors whether the error is a retryable
	retreyableErrors = map[string]bool{
		"Unavailable":         true,
		"InternalServerError": true,
	}

	// fcmServerUrl for testing purposes
	fcmServerUrl = fcm_server_url
)

// FcmClient stores the key and the Message (FcmMsg)
type FcmClient struct {
	ApiKey  string
	Message FcmMsg
}

// FcmMsg represents fcm request message
type FcmMsg struct {
	Data                  map[string]string   `json:"data,omitempty"`
	To                    string              `json:"to,omitempty"`
	RegistrationIds       []string            `json:"registration_ids,omitempty"`
	CollapseKey           string              `json:"collapse_key,omitempty"`
	Priority              string              `json:"priority,omitempty"`
	Notification          NotificationPayload `json:"notification,omitempty"`
	ContentAvailable      bool                `json:"content_available,omitempty"`
	DelayWhileIdle        bool                `json:"delay_while_idle,omitempty"`
	TimeToLive            int                 `json:"time_to_live,omitempty"`
	RestrictedPackageName string              `json:"restricted_package_name,omitempty"`
	DryRun                bool                `json:"dry_run,omitempty"`
	Condition             string              `json:"condition,omitempty"`
}

// FcmMsg represents fcm response message - (tokens and topics)
type FcmResponseStatus struct {
	Ok            bool
	StatusCode    int
	MulticastId   int                 `json:"multicast_id"`
	Success       int                 `json:"success"`
	Fail          int                 `json:"failure"`
	Canonical_ids int                 `json:"canonical_ids"`
	Results       []map[string]string `json:"results,omitempty"`
	MsgId         int                 `json:"message_id,omitempty"`
	Err           string              `json:"error,omitempty"`
	RetryAfter    string
}

// NotificationPayload notification message payload
type NotificationPayload struct {
	Title        string `json:"title,omitempty"`
	Body         string `json:"body,omitempty"`
	Icon         string `json:"icon,omitempty"`
	Sound        string `json:"sound,omitempty"`
	Badge        string `json:"badge,omitempty"`
	Tag          string `json:"tag,omitempty"`
	Color        string `json:"color,omitempty"`
	ClickAction  string `json:"click_action,omitempty"`
	BodyLocKey   string `json:"body_loc_key,omitempty"`
	BodyLocArgs  string `json:"body_loc_args,omitempty"`
	TitleLocKey  string `json:"title_loc_key,omitempty"`
	TitleLocArgs string `json:"title_loc_args,omitempty"`
}


// NewFcmClient init and create fcm client
func NewFcmClient(apiKey string) *FcmClient {
	fcmc := new(FcmClient)
	fcmc.ApiKey = apiKey

	return fcmc
}

// NewFcmTopicMsg sets the targeted token/topic and the data payload
func (this *FcmClient) NewFcmTopicMsg(to string, body map[string]string) *FcmClient {

	this.NewFcmMsgTo(to, body)

	return this
}

// NewFcmMsgTo sets the targeted token/topic and the data payload
func (this *FcmClient) NewFcmMsgTo(to string, body map[string]string) *FcmClient {
	this.Message.To = to
	this.Message.Data = body

	return this
}

// SetMsgData sets data payload
func (this *FcmClient) SetMsgData(body map[string]string) *FcmClient {

	this.Message.Data = body

	return this

}

// NewFcmRegIdsMsg gets a list of devices with data payload
func (this *FcmClient) NewFcmRegIdsMsg(list []string, body map[string]string) *FcmClient {
	this.newDevicesList(list)
	this.Message.Data = body

	return this

}

// newDevicesList init the devices list
func (this *FcmClient) newDevicesList(list []string) *FcmClient {
	this.Message.RegistrationIds = make([]string, len(list))
	copy(this.Message.RegistrationIds, list)

	return this

}

// AppendDevices adds more devices/tokens to the Fcm request
func (this *FcmClient) AppendDevices(list []string) *FcmClient {

	this.Message.RegistrationIds = append(this.Message.RegistrationIds, list...)

	return this
}

// apiKeyHeader generates the value of the Authorization key
func (this *FcmClient) apiKeyHeader() string {
	return fmt.Sprintf("key=%v", this.ApiKey)
}

// sendOnce send a single request to fcm
func (this *FcmClient) sendOnce() (*FcmResponseStatus, error) {

	fcmRespStatus := new(FcmResponseStatus)

	bodyBuff, _ := json.Marshal(this.Message)

	request, err := http.NewRequest("POST", fcmServerUrl, bytes.NewBuffer(bodyBuff))
	request.Header.Set("Authorization", this.apiKeyHeader())
	request.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	response, err := client.Do(request)

	if err != nil {
		fcmRespStatus.Ok = false
		return fcmRespStatus, err
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)

	fcmRespStatus.StatusCode = response.StatusCode

	fcmRespStatus.RetryAfter = response.Header.Get(retry_after_header)

	if response.StatusCode == 200 && err == nil {

		fmt.Print("came to success");
		fcmRespStatus.Ok = true

		eror := fcmRespStatus.parseStatusBody(body)
		if eror != nil {
			return fcmRespStatus, eror
		}

		return fcmRespStatus, nil

	} else {

		fmt.Print("came to failure", response.StatusCode);
		fcmRespStatus.Ok = false

		eror := fcmRespStatus.parseStatusBody(body)
		if eror != nil {
			return fcmRespStatus, eror
		}

		return fcmRespStatus, err
	}

	return fcmRespStatus, nil

}

// Send to fcm
func (this *FcmClient) Send() (*FcmResponseStatus, error) {
	return this.sendOnce()

}

// toJsonByte converts FcmMsg to a json byte
func (this *FcmMsg) toJsonByte() ([]byte, error) {

	return json.Marshal(this)

}
// parseStatusBody parse FCM response body
func (this *FcmResponseStatus) parseStatusBody(body []byte) error {

	if err := json.Unmarshal([]byte(body), &this); err != nil {
		return err
	}
	return nil

}

// SetPriorety Sets the priority of the message.
// Priority_HIGH or Priority_NORMAL
func (this *FcmClient) SetPriorety(p string) {

	if p == Priority_HIGH {
		this.Message.Priority = Priority_HIGH
	} else {
		this.Message.Priority = Priority_NORMAL
	}
}

// SetCollapseKey This parameter identifies a group of messages
// (e.g., with collapse_key: "Updates Available") that can be collapsed,
// so that only the last message gets sent when delivery can be resumed.
// This is intended to avoid sending too many of the same messages when the
// device comes back online or becomes active (see delay_while_idle).
func (this *FcmClient) SetCollapseKey(val string) *FcmClient {

	this.Message.CollapseKey = val

	return this
}

// SetNotificationPayload sets the notification payload based on the specs
// https://firebase.google.com/docs/cloud-messaging/http-server-ref
func (this *FcmClient) SetNotificationPayload(payload *NotificationPayload) *FcmClient {

	this.Message.Notification = *payload

	return this
}

// SetContentAvailable On iOS, use this field to represent content-available
// in the APNS payload. When a notification or message is sent and this is set
// to true, an inactive client app is awoken. On Android, data messages wake
// the app by default. On Chrome, currently not supported.
func (this *FcmClient) SetContentAvailable(isContentAvailable bool) *FcmClient {

	this.Message.ContentAvailable = isContentAvailable

	return this
}

// SetDelayWhileIdle When this parameter is set to true, it indicates that
// the message should not be sent until the device becomes active.
// The default value is false.
func (this *FcmClient) SetDelayWhileIdle(isDelayWhileIdle bool) *FcmClient {

	this.Message.DelayWhileIdle = isDelayWhileIdle

	return this
}

// SetTimeToLive This parameter specifies how long (in seconds) the message
// should be kept in FCM storage if the device is offline. The maximum time
// to live supported is 4 weeks, and the default value is 4 weeks.
// For more information, see
// https://firebase.google.com/docs/cloud-messaging/concept-options#ttl
func (this *FcmClient) SetTimeToLive(ttl int) *FcmClient {

	if ttl > MAX_TTL {

		this.Message.TimeToLive = MAX_TTL

	} else {

		this.Message.TimeToLive = ttl

	}
	return this
}

// SetRestrictedPackageName This parameter specifies the package name of the
// application where the registration tokens must match in order to
// receive the message.
func (this *FcmClient) SetRestrictedPackageName(pkg string) *FcmClient {

	this.Message.RestrictedPackageName = pkg

	return this
}

// SetDryRun This parameter, when set to true, allows developers to test
// a request without actually sending a message.
// The default value is false
func (this *FcmClient) SetDryRun(drun bool) *FcmClient {

	this.Message.DryRun = drun

	return this
}

// PrintResults prints the FcmResponseStatus results for fast using and debugging
func (this *FcmResponseStatus) PrintResults() {
	fmt.Println("Status Code   :", this.StatusCode)
	fmt.Println("Success       :", this.Success)
	fmt.Println("Fail          :", this.Fail)
	fmt.Println("Canonical_ids :", this.Canonical_ids)
	fmt.Println("Topic MsgId   :", this.MsgId)
	fmt.Println("Topic Err     :", this.Err)
	for i, val := range this.Results {
		fmt.Printf("Result(%d)> \n", i)
		for k, v := range val {
			fmt.Println("\t", k, " : ", v)
		}
	}
}

// IsTimeout check whether the response timeout based on http response status
// code and if any error is retryable
func (this *FcmResponseStatus) IsTimeout() bool {
	if this.StatusCode > 500 {
		return true
	} else if this.StatusCode == 200 {
		for _, val := range this.Results {
			for k, v := range val {
				if k == error_key && retreyableErrors[v] == true {
					return true
				}
			}
		}
	}

	return false
}

// GetRetryAfterTime converts the retrey after response header
// to a time.Duration
func (this *FcmResponseStatus) GetRetryAfterTime() (t time.Duration, e error) {
	t, e = time.ParseDuration(this.RetryAfter)
	return
}

// SetCondition to set a logical expression of conditions that determine the message target
func (this *FcmClient) SetCondition(condition string) *FcmClient {
	this.Message.Condition = condition
	return this
}

func SendUpvoteNotification(userId string, postId string, postUserId string, votes int, postType string, content string) {

	if (postUserId == userId) {
		return;
	}

	token, err := findUserById(postUserId);

	if (err != nil) {
		log.Print(err.Error())
	}

	notificationData := map[string]string{
		"id": strconv.Itoa(rand.Intn(999999)),
		"postId": postId,
		"time": time.Now().UTC().Format("2006-01-02T15:04:05.000Z"),
		"emphasis" : strconv.Itoa(votes) + " upvotes",
	}
	notificationData["type"] = postType;

	if (len(content) > 55) {
		content = content[0:55]
		content += "..."
	} else {
		content = content
	}

	var message string
	if (len(content) == 0) {
		message = "You got " + strconv.Itoa(votes) + " upvotes for your post";
	} else {
		message = "You got " + strconv.Itoa(votes) + " upvotes for your post \"" + content + "\"";
	}

	ids := []string{token.FcmToken}

	notificationData["message"] = message;

	sendNotification(ids, notificationData);
}

func SendCommentUpvoteNotification(userId string, commentUserId string,commentId string, postId string, commentType string, votes int, content string) {

	if (commentUserId == userId) {
		return;
	}

	token, err := findUserById(commentUserId);
	if (err != nil) {
		return;
	}

	data := map[string]string{
		"id": strconv.Itoa(rand.Intn(999999)),
		"postId": postId,
		"commentId": commentId,
		"time": time.Now().UTC().Format("2006-01-02T15:04:05.000Z"),
		"emphasis": strconv.Itoa(votes) + " upvotes",
		"type": commentType,
	}

	if (len(content) > 55) {
		content = content[0:55]
		content += "..."
	} else {
		content = content
	}

	message := "You got " + strconv.Itoa(votes) + " upvotes for your comment \"" + content + "\"";

	data["message"] = message;

	ids := []string{token.FcmToken}

	sendNotification(ids, data);
}

func findUserById(userId string) (user models.User, err error) {
	context := NewContext()
	defer context.Close()
	c := context.DbCollection("users")

	err = c.FindId(bson.ObjectIdHex(userId)).One(&user);
	return
}

func GetCommentsForPost(postId string) (comments []models.Comment, err error) {
	context := NewContext()
	defer context.Close()
	c := context.DbCollection("comments")

	err = c.Find(bson.M{"postId": bson.ObjectIdHex(postId)}).All(&comments)

	if (comments == nil) {
		comments = []models.Comment{}
	}

	return
}

func findNearByUsers(lat float64, lng float64) (users []models.User, err error) {
	context := NewContext()
	defer context.Close()
	c := context.DbCollection("users")

	now := time.Now().UTC()
	then := now.AddDate(0, 0, -10)

	currentLatLng := [2]float64{lng, lat}
	err = c.Find(bson.M{"loc": bson.M{"$near": currentLatLng}, "isActive" : true, "modifiedOn": bson.M{"$gt": then}}).Limit(30).All(&users);
	if (users == nil) {
		users = []models.User{}
	}
	return
}

func SendCommentNotification(postUserId string, postId string, commentUserId string, commentId string, postContent string,commentContent string, commentUsername string, commentJid string, commentType string) {

	user, err := findUserById(postUserId);

	if (err != nil) {
		return;
	}

	comments, e := GetCommentsForPost(postId)

	if (e != nil) {
		return;
	}

	var userIds []bson.ObjectId;
	for i := 0; i < len(comments); i++ {
		userIds = append(userIds, comments[i].UserId)
	}

	options := bson.M{"_id" : bson.M{"$in" : userIds}}

	context := NewContext()
	defer context.Close()
	c := context.DbCollection("users")

	var users []models.User
	err = c.Find(options).All(&users);

	var fcmIds []string

	for i := 0; i < len(users); i++ {
		if (users[i].Id.Hex() != commentUserId && users[i].IsActive) {
			fcmIds = append(fcmIds, users[i].FcmToken)
		}
	}

	notificationData := map[string]string{
		"id": strconv.Itoa(rand.Intn(999999)),
		"postId": postId,
		"commentId": commentId,
		"jid": commentJid,
		"content": commentContent,
		"username": commentUsername,
		"userId": commentUserId,
		"time": time.Now().UTC().Format("2006-01-02T15:04:05.000Z"),
		"type": commentType,
	}

	var content string;

	if (len(postContent) > 25) {
		content = postContent[0:25]
		content += "..."
	} else {
		content = postContent
	}

	if (len(commentContent) > 30) {
		commentContent = commentContent[0:30]
		commentContent += "..."
	} else {
		commentContent = commentContent
	}

	message := "";

	if (len(postContent) == 0) {
		message = "Someone else commented \"" + commentContent + "\"";
	} else {
		message = "Someone else commented \"" + commentContent + "\"" + " on the post \"" + content + "\"";
	}

	notificationData["message"] = message;

	if (len(fcmIds) > 0) {
		sendNotification(fcmIds, notificationData);
	}

	found := false;
	for i := 0; i < len(users); i++ {
		if (users[i].Id.Hex() == user.Id.Hex()) {
			found = true;
		}
	}

	if (found || postUserId == commentUserId) {
		return;
	}

	ids := []string{user.FcmToken}

	message = "Someone commented \"" + commentContent + "\"" + " on your post \"" + content + "\"";

	notificationData["message"] = message;

	sendNotification(ids, notificationData);
}

func SendNearByNotification(post models.Post) {
	users, err := findNearByUsers(post.Loc[1], post.Loc[0]);

	if (err != nil) {
		return;
	}

	ids := []string{}

	for _, user := range users {
		if (user.Id.Hex() != post.UserId.Hex()) {
			ids = append(ids, user.FcmToken);
		}
	}

	if (len(ids) == 0) {
		return;
	}

	data := map[string]string{
		"id": strconv.Itoa(rand.Intn(999999)),
		"postId": post.Id.Hex(),
		"time": time.Now().UTC().Format("2006-01-02T15:04:05.000Z"),
	}

	if (len(post.ImageUrl) > 0) {
		data["type"] = "image_post";
	} else {
		data["type"] = "post";
	}

	var content string;

	if (len(post.Content) > 80) {
		content = post.Content[0:80]
		content += "..."
	} else {
		content = post.Content
	}

	var message string
	if (len(content) == 0) {
		message = "Someone nearby just posted";
	} else {
		message = "Someone nearby posted \"" + content + "\"";
	}

	data["message"] = message;

	sendNotification(ids, data);
}

func FeedbackNotification(fcmToken string) {
	data := map[string]string{
		"id": strconv.Itoa(rand.Intn(999999)),
		"time": time.Now().UTC().Format("2006-01-02T15:04:05.000Z"),
		"type": "general",
		"message": "Feedback succesfully submitted, thank you for your valuable feedback!!",
	}
	ids := []string{fcmToken}
	sendNotification(ids, data);
}

func SendDeletePostNotification(userId string, content string) {
	token, err := findUserById(userId);
	if (err != nil) {
		return;
	}

	data := map[string]string{
		"id": strconv.Itoa(rand.Intn(999999)),
		"time": time.Now().UTC().Format("2006-01-02T15:04:05.000Z"),
		"type": "general",
	}

	if (len(content) > 30) {
		content = content[0:30]
		content += "..."
	} else {
		content = content
	}

	message := "";
	if (len(content) == 0) {
		message = "Your Post has been suspended as people in your community downvoted or reported about it"
	} else {
		message = "Your Post \"" + content + "\" has been suspended as people in your community downvoted or reported about it"
	}
	ids := []string{token.FcmToken}

	data["message"] = message;

	sendNotification(ids, data);
}

func GroupUnlockedNotification(user models.User) {
	message := "You have unlocked your college/workplace, you can now share the happening with your college/workplace even when you're away"
	data := map[string]string{
		"id": strconv.Itoa(rand.Intn(999999)),
		"time": time.Now().UTC().Format("2006-01-02T15:04:05.000Z"),
		"type": "general",
	}

	ids := []string{user.FcmToken}

	data["message"] = message;

	sendNotification(ids, data);
}

func AppAvailableNotification(city string) {
	//TODO: We can have a look at this function later
	message := "This App is now available in your location!!!"
	context := NewContext()
	defer context.Close()
	c := context.DbCollection("users")

	var users []models.User
	err := c.Find(bson.M{"city": city, "IsActive" : false}).All(&users);

	if (err != nil) {
		return;
	}

	data := map[string]string{
		"id": strconv.Itoa(rand.Intn(999999)),
		"time": time.Now().UTC().Format("2006-01-02T15:04:05.000Z"),
		"type": "general",
	}

	var ids  []string;

	for _, user := range users {
		ids = append(ids, user.FcmToken)
	}

	data["message"] = message;

	sendNotification(ids, data);
}

func sendNotification(fcmTokens []string, data map[string]string) {

	c := NewFcmClient(serverKey)
	c.NewFcmRegIdsMsg(fcmTokens, data)

	_, err := c.Send()

	if (err != nil) {
		slack.Send(slack.ErrorLevel, "FCM Notification error: " + err.Error())
	}
}
