package main

import (
	"flag"
	"log"
	"phone_ios_backend/config"
	"phone_ios_backend/config/sms"
	"phone_ios_backend/connection"
	"phone_ios_backend/logger"
	"phone_ios_backend/router"

	"github.com/gin-gonic/gin"
)

func main() {
	configFile := flag.String("config", "./config/config.yaml", "path to the configuration file")
	flag.Parse()
	if err := config.LoadConfig(*configFile); err != nil {
		log.Fatal("loading config fatal!")
	}
	logger.NewZapLogger()
	_, err := connection.ConnectDatabase(config.Setting)
	if err != nil {
		log.Fatal("get db connection fatal!")
	}
	if err = sms.CreateSMSClient(); err != nil {
		log.Fatal("get sms client fatal!")
	}
	connection.InitRedis()
	g := gin.Default()
	router.Router(g)
	g.Run(":8089")
}
