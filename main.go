package main

import (
	"fmt"
	"github.com/mymachine8/fardo-api/controllers"
	"net/http"
	"github.com/facebookgo/grace/gracehttp"
	"github.com/mymachine8/fardo-api/common"
)

func main() {

	common.StartUp();
	handler := controllers.InitRoutes()
	var port string = "8082"//os.Getenv("PORT")
	fmt.Println("Starting server on :8082");

	err := gracehttp.Serve(
		&http.Server{Addr: "localhost:" + port, Handler: handler});

	if(err != nil) {
		println(err.Error());
	}
}