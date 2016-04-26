package goros

import (
	"encoding/json"
	"fmt"
)

type Topic struct {
	Op    string          `json:"op"`
	Id    string          `json:"id"`
	Topic string          `json:"topic"`
	Msg   json.RawMessage `json:"msg,omitempty"`
}

type TopicCallback func(*json.RawMessage)

func NewTopic(topicName string) *Topic {
	topic := &Topic{Op: "subscribe", Topic: topicName}
	topic.Id = fmt.Sprintf("%s:%s:%d", topic.Op, topic.Topic, messageCount)
	messageCount++

	return topic
}
