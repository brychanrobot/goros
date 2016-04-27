package main

import (
	"encoding/json"
	"flag"
	"log"

	"github.com/brychanrobot/goros"
)

type Uav struct {
	Tracks []Track `json:"tracks"`
}

type Track struct {
	MeanX       float64 `json:"meanx"`
	MeanY       float64 `json:"meany"`
	Label       int     `json:"label"`
	InlierRatio float64 `json:"inliers"`
}

func main() {
	flag.Parse()
	ros := goros.NewRos("ws://192.168.27.201:9090")

	topics := ros.GetTopics()
	log.Println(topics)

	services := ros.GetServices()
	log.Println(services)

	ros.Subscribe("/visual_mtt/rransac_tracks", func(msg *json.RawMessage) {
		var uav Uav
		json.Unmarshal(*msg, &uav)
		log.Println(uav)
	})

	select {} //keeps the application open
}
