package main

import (
	"fmt"
	"log"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
)

func main() {
	c_string := "amqp://guest:guest@localhost:5672/"

	fmt.Println("Starting Peril client...")

	conn, err := amqp.Dial(c_string)
	if err != nil {
		log.Fatal("Couldn't connect to server:", err)
	}
	defer conn.Close()

	c_name, err := gamelogic.ClientWelcome()
	if err != nil {
		log.Fatal("Couldn't finish ClientWelcome:", err)
	}

	state := gamelogic.NewGameState(c_name)
	pause_handler := handlerPause(state)
	exchange, queue_name, queue_type := routing.ExchangePerilDirect, fmt.Sprintf("%v.%v", routing.PauseKey, c_name), pubsub.QueueTypeTransient
	if err := pubsub.SubscribeJSON(conn, exchange, queue_name, routing.PauseKey, queue_type, pause_handler); err != nil {
		log.Fatal("Couldn't subscribe to new message queue:", err)
	}

	for {
		inp := gamelogic.GetInput()
		if inp[0] == "spawn" {
			if err := state.CommandSpawn(inp); err != nil {
				fmt.Println(err)
			}
		} else if inp[0] == "move" {
			if _, err := state.CommandMove(inp); err != nil {
				fmt.Println(err)
			}
		} else if inp[0] == "status" {
			state.CommandStatus()
		} else if inp[0] == "help" {
			gamelogic.PrintClientHelp()
		} else if inp[0] == "spam" {
			fmt.Println("Spamming not allowed yet!")
		} else if inp[0] == "quit" {
			gamelogic.PrintQuit()
			break
		} else {
			fmt.Println("Unknown command. For help write \"help\"")
		}
	}
}
