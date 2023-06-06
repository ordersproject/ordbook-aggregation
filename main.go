package main

import (
	"flag"
	"fmt"
	"ordbook-aggregation/conf"
	"ordbook-aggregation/config"
	"ordbook-aggregation/controller"
	"ordbook-aggregation/major"
	_ "ordbook-aggregation/service/cache_service"
	"ordbook-aggregation/ws_service/ws"
)

var ENV string

func init() {
	flag.StringVar(&ENV, "env", "example", "EnvironmentEnum")
}

func InitEnv() {
	flag.Parse()
	if ENV == "example" {
		conf.SystemEnvironmentEnum = conf.ExampleEnvironmentEnum
	}
	fmt.Println(fmt.Sprintf("%s%v", "Env : ", ENV))
}


func InitAll() {
	logName := "ordbook-aggregation"
	major.InitLogger(logName)
	config.InitConfig()
	major.InitMongo()
}

func run() {
	var(
		endRunning = make(chan bool, 1)
	)
	<- endRunning
}

// @title OrdBook API Service
// @version 1.0
// @description  OrdBook API Service
// @termsOfService
// @contact.name API Support
// @schemes https
// @BasePath /book
func main() {
	InitEnv()
	InitAll()

	go ws.StartWS()

	controller.Run()
	//run()
}
