package controllers

import (
	"github.com/julienschmidt/httprouter"
	"github.com/google/go-gcm"
	"net/http"
	"fmt"
	"github.com/mymachine8/fardo-api/models"
	"gopkg.in/mgo.v2/bson"
	"encoding/json"
	"log"
	"github.com/urfave/negroni"
	"github.com/mymachine8/fardo-api/data"
	"github.com/mymachine8/fardo-api/common"
)


func InitRoutes() *httprouter.Router{
	r := httprouter.New();

	n := negroni.New();
	n.Use(negroni.HandlerFunc(common.Authorize));
	n.UseHandler(r);

	r.GET("/", helloWorldHandler);

	//TODO: Other routers come here
	r.GET("/my-circle", myCircleHandler);


	r.GET("/category", categoryListHandler);
	r.GET("/group", groupListHandler);
	r.GET("/group/:id/label", labelListHandler);
	r.POST("/group", createGroupHandler);
	r.POST("/label/bulk", createLabelsBulkHandler);
	r.DELETE("/label/bulk", removeLabelsBulkHandler);

	r.POST("/user", createUserHandler);

	r.POST("/post", createPostHandler);

	return r;
}

func helloWorldHandler(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
	fmt.Fprintln(rw, "Hello World")
}

func myCircleHandler(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
	v := r.URL.Query()
	fmt.Println(v["lat"]);
	var latLong [2]float64 = [2] float64{3232.323,32323.3232};
	result := models.MyPosts(bson.NewObjectId(),latLong);
	rw.Header().Set("Content-Type", "application/json")

	var error models.FardoError
	response := struct {
		Data [] models.Post `json:"data"`
		Error models.FardoError `json:"error,omitempty"`
	} {
		result,
		error,
	}

	apiKey := "AIzaSyDcXHtO1YB-tCRwFAvaVqJDud8gR00VJs4"
	m := gcm.HttpMessage{};
	m.To = "euF-S2d-RFk:APA91bHMBhCCHQVJ0uAeLOHiF9v3azYnzTn1tLx8fz7u2_7nCQMExJt6QZxJLeEQK4W1yrSi0vqCLtEVPOhQwHnXbkRKUZRunXDTz1ZOZKi4F5RrdWN-JRvgyFg2nV6yiwhZ2eWQbd2y" //Registration Token of the user
	m.CollapseKey = "" //A group tag which specifies all the messages send are of the same type, so that server can send recent msg when the device is awake
	m.Priority = "" //Valid values "normal"/ "high"
	m.DelayWhileIdle = true //When this parameter is set to true, it indicates that the message should not be sent until the device becomes active.
	// m.TimeToLive 0 to 2,419,200 seconds
	x := map[string]interface{}{
		"bar": "foo",
		"baz": "hello world",
	}
	//gcmMsg := models.GcmMessage {"Hello World", "GCM"}
	m.Data = x//This should be flat struct with strings, Values in string types are recommended. You have to convert values in objects or other non-string data types (e.g., integers or booleans) to string.
	//m.Notification will be sent as notification to the android device, to the notifications tray

	gcmResponse, e := gcm.SendHttp(apiKey, m);

	if(e != nil) {
		println(e.Error())
	}

	println(gcmResponse.Failure);
	println(gcmResponse.Success);

	jsonResult, _ := json.Marshal(response);

	rw.Write(jsonResult)
}

func createUserHandler(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
	//TODO:Create User Do all necessary validations and save it to DB
	decoder := json.NewDecoder(r.Body)
	var user models.User
	err := decoder.Decode(&user);
	if err != nil {
		panic(err)
	}
	log.Println(user.Imei)
}

func createPostHandler(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
	//TODO:Create Post Do all necessary validations and save it to DB
	decoder := json.NewDecoder(r.Body)
	var post models.Post
	err := decoder.Decode(&post);
	if err != nil {
		panic(err)
	}
	log.Println(post)
}

func createGroupHandler(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {

}

func createLabelsBulkHandler(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
	//TODO:Create Post Do all necessary validations and save it to DB
	decoder := json.NewDecoder(r.Body)
	var post models.Post
	err := decoder.Decode(&post);
	if err != nil {
		panic(err)
	}
	log.Println(post)
}


func removeLabelsBulkHandler(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
	//TODO:Create Post Do all necessary validations and save it to DB
	decoder := json.NewDecoder(r.Body)
	var post models.Post
	err := decoder.Decode(&post);
	if err != nil {
		panic(err)
	}
	log.Println(post)
}

func categoryListHandler(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {

	result := data.GetAllCategories();
	var error models.FardoError
	response := struct {
		Data interface{} `json:"data"`
		Error models.FardoError `json:"error,omitempty"`
	} {
		result,
		error,
	}

	jsonResult, _ := json.Marshal(response);

	rw.Write(jsonResult)

}

func groupListHandler(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
	//TODO:Create Post Do all necessary validations and save it to DB
	decoder := json.NewDecoder(r.Body)
	var post models.Post
	err := decoder.Decode(&post);
	if err != nil {
		panic(err)
	}
	log.Println(post)
}

func labelListHandler(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
	//TODO:Create Post Do all necessary validations and save it to DB

}
