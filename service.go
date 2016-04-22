package goros

import (
	"encoding/json"
	"fmt"
)

type ServiceCall struct {
	Op      string `json:"op"`
	Id      string `json:"id"`
	Service string `json:"service"`
	Args    string `json:"args,omitempty"`
}

func newServiceCall(service string) *ServiceCall {
	serviceCall := &ServiceCall{Op: "call_service", Service: service}
	serviceCall.Id = fmt.Sprintf("%s:%s:%d", serviceCall.Op, serviceCall.Service, messageCount)
	messageCount++

	return serviceCall
}

type ServiceResponse struct {
	Op      string                     `json:"op"`
	Id      string                     `json:"id"`
	Service string                     `json:"service"`
	Result  bool                       `json:"result"`
	Values  map[string]json.RawMessage `json:"values"`
}
