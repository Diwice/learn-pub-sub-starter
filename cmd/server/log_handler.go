package main

import (
	"fmt"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
)

func handlerLogs() func(routing.GameLog) pubsub.AckType {
	return func(rgm routing.GameLog) pubsub.AckType {
		defer fmt.Print("> ")
		if err := gamelogic.WriteLog(rgm); err != nil {
			return pubsub.NackRequeue
		}
		return pubsub.Ack
	}
}
