package cron

import (
	"github.com/mymachine8/fardo-api/data"
	"log"
	"github.com/robfig/cron"
)


func InitCronJobs() {
	c := cron.New()
	c.AddFunc("@every 6h", groupTrendingScore)
	c.Start()
}

func groupTrendingScore() {
	log.Print("computation started");
	err := data.CalculatePlacesTrendingScore();

	if(err!= nil) {
		log.Print(err.Error())
	}
}
