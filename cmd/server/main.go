package main

import (
	"os"
	"fmt"
	"log"
	"os/signal"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
)

func main() {
	c_string := "amqp://guest:guest@localhost:5672/"

	fmt.Println("Starting Peril server...")

	conn, err := amqp.Dial(c_string)
	if err != nil {
		log.Fatal("Couldn't connect to server:", err)
	}
	defer conn.Close()

	new_ch, err := conn.Channel()
	if err != nil {
		log.Fatal("Couldn't create a channel:", err)
	}

	fmt.Println("Successfully connected to the server")

	if err = pubsub.PublishJSON(new_ch, routing.ExchangePerilDirect, routing.PauseKey, routing.PlayingState{IsPaused: true}); err != nil {
		log.Fatal("Couldn't post the message:", err)
	}

	signal_ch := make(chan os.Signal, 1)
	signal.Notify(signal_ch, os.Interrupt)
	<-signal_ch

	fmt.Println("Application is shutting down...")
}
