package services

import "github.com/profzone/chain/messages"

type Service interface {
	Messages() []messages.MessageHandler
	Start() error
	Stop() error
}

type serviceConstructor func() Service

