package main

import (
	"fmt"
	"github.com/mymachine8/fardo-api/controllers"
	"github.com/mymachine8/fardo-api/bootstrap/dbconn"
	"net/http"
	"github.com/facebookgo/grace/gracehttp"
)

func main() {
	r := controllers.GetRouter()
	var port string = "8082"//os.Getenv("PORT")
	dbconn.GetInstance()
	fmt.Println("Starting server on :8082");
	err := gracehttp.Serve(
		&http.Server{Addr: "localhost:" + port, Handler: r});

	if(err != nil) {
		println(err.Error());
	}
}