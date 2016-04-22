package goros

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"regexp"
	"strings"
	"sync"

	"golang.org/x/net/websocket"
)

var (
	messageCount = 0
)

type Ros struct {
	origin           string
	url              string
	ws               *websocket.Conn
	receivedMapMutex sync.Mutex
	receivedMap      map[string]chan interface{}
}

func NewRos(url string) *Ros {
	ros := Ros{url: url, origin: "https://localhost"}
	ros.receivedMap = make(map[string]chan interface{})
	ros.connect()
	go ros.handleIncoming()
	return &ros
}

func (ros *Ros) connect() {
	ws, err := websocket.Dial(ros.url, "", ros.origin)
	if err != nil {
		log.Fatal(err)
	}

	ros.ws = ws
}

func (ros *Ros) getServiceResponse(service *ServiceCall) *ServiceResponse {
	response := make(chan interface{})
	ros.receivedMapMutex.Lock()
	ros.receivedMap[service.Id] = response
	ros.receivedMapMutex.Unlock()
	err := websocket.JSON.Send(ros.ws, service)
	if err != nil {
		fmt.Println("Couldn't send msg")
	}

	serviceResponse := <-response
	return serviceResponse.(*ServiceResponse)
}

func (ros *Ros) returnToAppropriateChannel(id string, data interface{}) {
	ros.receivedMapMutex.Lock()
	ros.receivedMap[id] <- data
	ros.receivedMapMutex.Unlock()
}

func (ros *Ros) handleIncoming() {
	var msg []byte
	for {
		err := websocket.Message.Receive(ros.ws, &msg)
		if err != nil {
			if err == io.EOF {
				break
			}
			fmt.Println("Couldn't receive msg " + err.Error())
			break
		}

		opRegex, err := regexp.Compile(`"op"\s*:\s*"[[:alpha:],_]*`)
		if err != nil {
			log.Println(err)
		}
		opString := opRegex.FindString(string(msg))
		splitOpString := strings.Split(opString, "\"")
		operation := splitOpString[len(splitOpString)-1]

		//log.Println(operation)

		/*
			var data map[string]interface{}
			jsonErr := json.Unmarshal(msg, &data)
			//fmt.Printf("Received from server: %s\n", data)
			if jsonErr != nil {
				panic(jsonErr)
			}

			ros.receivedMapMutex.Lock()
			ros.receivedMap[data["id"].(string)] <- data
			ros.receivedMapMutex.Unlock()
		*/
		if operation == "service_response" {
			var serviceResponse ServiceResponse
			json.Unmarshal(msg, &serviceResponse)
			ros.receivedMapMutex.Lock()
			ros.receivedMap[serviceResponse.Id] <- &serviceResponse
			ros.receivedMapMutex.Unlock()
		}
	}
}

func (ros *Ros) GetTopics() []string {
	response := ros.getServiceResponse(newServiceCall("/rosapi/topics"))
	var topics []string
	json.Unmarshal(response.Values["topics"], &topics)
	return topics
}

func (ros *Ros) Subscribe(topic *Topic) {

}
