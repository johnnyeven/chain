package services

import (
	"reflect"
	"fmt"
	"github.com/profzone/chain/messages"
)

type Stack struct {
	Name     string
	serviceFunc []serviceConstructor
	services map[reflect.Type]Service
}

var GeneralStack *Stack

func NewStack(name string) *Stack {
	if GeneralStack != nil {
		return GeneralStack
	}
	s := &Stack{
		Name:     name,
		services: make(map[reflect.Type]Service),
	}

	GeneralStack = s

	return s
}

func NewStackWithTemplate(presetName string) *Stack {
	if GeneralStack != nil {
		return GeneralStack
	}
	preset := presetStack[presetName]
	s := NewStack(presetName)

	for _, f := range preset {
		s.RegisterService(f)
	}

	return s
}

func (s *Stack) RegisterService(c serviceConstructor) {
	s.serviceFunc = append(s.serviceFunc, c)
}

func (s *Stack) GetService(serviceType reflect.Type) Service {
	return s.services[serviceType]
}

func (s *Stack) Start() error {
	for _, constructor := range s.serviceFunc {
		service := constructor()
		t := reflect.TypeOf(service)
		if t.Kind() == reflect.Ptr {
			v := reflect.ValueOf(service)
			if v.IsNil() {
				panic(fmt.Sprintf("service must not be nil, type: %v", t))
			}
		}
		focusMessage := service.Messages()
		for _, m := range focusMessage {
			messages.GetMessageManager().RegisterMessage(m)
		}
		s.services[t] = service
		service.Start()
	}
	return nil
}

func (s *Stack) Stop() error {
	for _, service := range s.services {
		service.Stop()
	}
	return nil
}
