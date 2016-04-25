package main

import (
	"encoding/json"
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
	ros := goros.NewRos("ws://192.168.27.20:9090")

	topics := ros.GetTopics()
	log.Println(topics)

	ros.Subscribe("/visual_mtt/rransac_tracks", func(msg *json.RawMessage) {
		var uav Uav
		json.Unmarshal(*msg, &uav)
		log.Println(uav)
	})

	select {} //keeps the application open
}
