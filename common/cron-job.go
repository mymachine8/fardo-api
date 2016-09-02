package common

import (
	"github.com/robfig/cron"
	"time"
	"log"
)

func InitCronJobs() {
	c := cron.New()
	log.Print(time.Now().UTC())
	log.Print(time.ParseDuration("48h"))
	//c.AddFunc("@every 48h", resetGroupTrendingScore)
	c.Start()
}
