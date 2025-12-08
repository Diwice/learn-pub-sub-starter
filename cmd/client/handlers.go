package main

import (
	"fmt"
	"time"
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

func handlerWar(gs *gamelogic.GameState, chn *amqp.Channel) func(gamelogic.RecognitionOfWar) pubsub.AckType {
	return func(grow gamelogic.RecognitionOfWar) pubsub.AckType {
		defer fmt.Print("> ")

		initiator, player := grow.Attacker.Username, gs.GetPlayerSnap().Username
		outcome, winner, loser := gs.HandleWar(grow)

		switch outcome {
		case gamelogic.WarOutcomeNotInvolved:
			return pubsub.NackRequeue
		case gamelogic.WarOutcomeNoUnits:
			return pubsub.Ack
		case gamelogic.WarOutcomeOpponentWon:
			if err := publishLog(chn, fmt.Sprintf("%v won a war against %v", winner, loser), initiator, player); err != nil {
				return pubsub.NackRequeue
			}
			return pubsub.Ack
		case gamelogic.WarOutcomeYouWon:
			if err := publishLog(chn, fmt.Sprintf("%v won a war against %v", winner, loser), initiator, player); err != nil {
				return pubsub.NackRequeue
			}
			return pubsub.Ack
		case gamelogic.WarOutcomeDraw:
			if err := publishLog(chn, fmt.Sprintf("A war between %v and %v resulted in a draw", winner, loser), initiator, player); err != nil {
				return pubsub.NackRequeue
			}
			return pubsub.Ack
		}

		fmt.Println("Unknown war outcome!")

		return pubsub.NackDiscard
	}
}

func publishLog(chn *amqp.Channel, msg, initiator, username string) error {
	exch, route_key, data := routing.ExchangePerilTopic, fmt.Sprintf("%v.%v", routing.GameLogSlug, initiator), routing.GameLog{CurrentTime: time.Now(), Message: msg, Username: username}
	if err := pubsub.PublishGob(chn, exch, route_key, data); err != nil {
		return err
	}

	return nil
}
