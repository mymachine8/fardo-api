package main

import (
	"fmt"
	"github.com/mymachine8/fardo-api/controllers"
	"github.com/mymachine8/fardo-api/bootstrap/dbconn"
	"net/http"
	"github.com/facebookgo/grace/gracehttp"
	"os"
)

func main() {
	r := controllers.GetRouter()
	var port string = os.Getenv("PORT")
	dbconn.GetInstance()
	fmt.Println("Starting server on :8082");
	gracehttp.Serve(
		&http.Server{Addr: ":" + port, Handler: r});
}