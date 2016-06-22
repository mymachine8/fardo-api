package fardoapp

import (
	"github.com/mymachine8/fardo-api/common"
	"github.com/mymachine8/fardo-api/controllers"
)

func Run() {
	common.StartUp();
	controllers.InitRoutes()
}