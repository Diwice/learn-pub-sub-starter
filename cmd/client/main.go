package main

import (
	"os"
	"fmt"
	"log"
	"os/signal"
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

	exchange, queue_name, queue_type := "peril_direct", fmt.Sprintf("%v.%v", routing.PauseKey, c_name), pubsub.QueueTypeTransient

	_, _, err = pubsub.DeclareAndBind(conn, exchange, queue_name, routing.PauseKey, queue_type)
	if err != nil {
		log.Fatal("Couldn't declare and bind a new queue:", err)
	}

	signal_ch := make(chan os.Signal, 1)
	signal.Notify(signal_ch, os.Interrupt)
	<-signal_ch

	fmt.Println("Application is shutting down...")
}
