package main

import (
	"fmt"
	"log"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
)

func send_msg(ch *amqp.Channel, play_state routing.PlayingState) error {
	if err := pubsub.PublishJSON(ch, routing.ExchangePerilDirect, routing.PauseKey, play_state); err != nil {
		return err
	}

	return nil
}

func main() {
	c_string := "amqp://guest:guest@localhost:5672/"

	fmt.Println("Starting Peril server...")

	conn, err := amqp.Dial(c_string)
	if err != nil {
		log.Fatal("Couldn't connect to server:", err)
	}
	defer conn.Close()

	new_ch, _, err := pubsub.DeclareAndBind(conn, routing.ExchangePerilTopic, routing.GameLogSlug, "game_logs.*", pubsub.QueueTypeDurable)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Successfully connected to the server")

	gamelogic.PrintServerHelp()

	for {
		inp := gamelogic.GetInput()
		if inp[0] == "pause" {
			fmt.Println("Sending pause message")
			if err := send_msg(new_ch, routing.PlayingState{IsPaused: true}); err != nil {
				log.Fatal(err)
			}
			fmt.Println("Sent pause message successfully")
		} else if inp[0] == "resume" {
			fmt.Println("Sending resume message")
			if err := send_msg(new_ch, routing.PlayingState{IsPaused: false}); err != nil {
				log.Fatal(err)
			}
			fmt.Println("Sent resume mesage successfully")
		} else if inp[0] == "quit" {
			fmt.Println("Quitting...")
			break
		} else {
			fmt.Println("Unknown command")
		}
	}
}
