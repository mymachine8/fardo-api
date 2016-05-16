package controllers

import (
	"github.com/julienschmidt/httprouter"
	"net/http"
	"fmt"
	"github.com/mymachine8/fardo-api/models"
	"gopkg.in/mgo.v2/bson"
	"encoding/json"
)

func GetRouter() *httprouter.Router{
	r := httprouter.New();
	r.GET("/", helloWorldHandler);

	//TODO: Other routers come here
	r.GET("/my-circle", myCircleHandler);

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
	jsonResult, _ := json.Marshal(result);
	rw.Write(jsonResult)
}
