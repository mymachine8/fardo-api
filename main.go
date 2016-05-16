package main

import (
	"fmt"
	"github.com/mymachine8/fardo-api/controllers"
	"github.com/mymachine8/fardo-api/bootstrap/dbconn"
	"net/http"
	"github.com/facebookgo/grace/gracehttp"
)

var (
	address0 = "localhost:8082"
)
func main() {
	r := controllers.GetRouter()
	dbconn.GetInstance()
	fmt.Println("Starting server on :8082");
	gracehttp.Serve(
		&http.Server{Addr: address0, Handler: r});
}