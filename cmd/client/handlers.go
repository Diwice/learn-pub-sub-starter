package main

import (
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
)

func handlerPause(gs *gamelogic.GameState) func(routing.PlayingState) pubsub.AckType {
	return func(rps routing.PlayingState) pubsub.AckType {
		defer fmt.Print("> ")
		gs.HandlePause(rps)
		return pubsub.Ack
	}
}

func handlerMove(gs *gamelogic.GameState, chn *amqp.Channel) func(gamelogic.ArmyMove) pubsub.AckType {
	return func(gam gamelogic.ArmyMove) pubsub.AckType {
		defer fmt.Print("> ")
		outcome := gs.HandleMove(gam)

		switch outcome {
		case gamelogic.MoveOutComeSafe:
			return pubsub.Ack
		case gamelogic.MoveOutcomeMakeWar:
			c_key := fmt.Sprintf("%v.%v", routing.WarRecognitionsPrefix, gs.GetUsername())
			rec_of_war := gamelogic.RecognitionOfWar{Attacker: gam.Player, Defender: gs.GetPlayerSnap()}
			if err := pubsub.PublishJSON(chn, routing.ExchangePerilTopic, c_key, rec_of_war); err != nil {
				return pubsub.NackRequeue
			}

			return pubsub.Ack
		}

		return pubsub.NackDiscard
	}
}

func handlerWar(gs *gamelogic.GameState) func(gamelogic.RecognitionOfWar) pubsub.AckType {
	return func(grow gamelogic.RecognitionOfWar) pubsub.AckType {
		defer fmt.Print("> ")

		outcome, _, _ := gs.HandleWar(grow)

		switch outcome {
		case gamelogic.WarOutcomeNotInvolved:
			return pubsub.NackRequeue
		case gamelogic.WarOutcomeNoUnits:
			return pubsub.Ack
		case gamelogic.WarOutcomeOpponentWon:
			return pubsub.Ack
		case gamelogic.WarOutcomeYouWon:
			return pubsub.Ack
		case gamelogic.WarOutcomeDraw:
			return pubsub.Ack
		}

		fmt.Println("Unknown war outcome!")

		return pubsub.NackDiscard
	}
}
