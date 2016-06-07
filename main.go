package main

import (
	"fmt"
	"github.com/mymachine8/fardo-api/controllers"
	"net/http"
	"github.com/facebookgo/grace/gracehttp"
	"github.com/mymachine8/fardo-api/common"
	"github.com/rs/cors"
)

func main() {

	common.StartUp();
	n := controllers.InitRoutes()
	var port string = "8082"//os.Getenv("PORT")
	fmt.Println("Starting server on :8082");

	c := cors.New(cors.Options{
		AllowCredentials: true,
		Debug: true,
		AllowedMethods : []string{"GET", "POST","PUT", "DELETE"},
		AllowedHeaders :[]string{"Origin", "Accept", "Content-Type", "Authorization"},
	})
	handler := c.Handler(n)
	err := gracehttp.Serve(
		&http.Server{Addr: "localhost:" + port, Handler: handler});

	if(err != nil) {
		println(err.Error());
	}
}