package cron

import (
	"github.com/mymachine8/fardo-api/data"
	"log"
)


func GroupTrendingScore() {
	err := data.CalcuatePlacesTrendingScore();

	if(err!= nil) {
		log.Print(err.Error())
	}
}
