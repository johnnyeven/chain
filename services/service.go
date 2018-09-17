package services

import "git.profzone.net/profzone/chain/messages"

type Service interface {
	Messages() []messages.MessageHandler
	Start() error
	Stop() error
}

type serviceConstructor func() Service

