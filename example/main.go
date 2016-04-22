package main

import (
	"log"

	"github.com/brychanrobot/goros"
)

func main() {
	ros := goros.NewRos("ws://192.168.27.20:9090")

	topics := ros.GetTopics()
	log.Println(topics)

	select {} //keeps the application open
}
