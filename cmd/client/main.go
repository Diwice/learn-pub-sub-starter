package main

import (
	"fmt"
	"log"
	"time"
	"strconv"
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

	new_ch, err := conn.Channel()
	if err != nil {
		log.Fatal("Couldn't create a move channel:", err)
	}

	state := gamelogic.NewGameState(c_name)
	pause_handler := handlerPause(state)
	exchange, queue_name, queue_type := routing.ExchangePerilDirect, fmt.Sprintf("%v.%v", routing.PauseKey, c_name), pubsub.QueueTypeTransient
	if err := pubsub.SubscribeJSON(conn, exchange, queue_name, routing.PauseKey, queue_type, pause_handler); err != nil {
		log.Fatal("Couldn't subscribe to the pause queue:", err)
	}

	move_handler := handlerMove(state, new_ch)
	move_exch, move_q_name, move_key, move_q_type := routing.ExchangePerilTopic, fmt.Sprintf("%v.%v", "army_moves", c_name), "army_moves.*", pubsub.QueueTypeTransient
	if err := pubsub.SubscribeJSON(conn, move_exch, move_q_name, move_key, move_q_type, move_handler); err != nil {
		log.Fatal("Couldn't subscribe to the move queue:", err)
	}

	war_handler := handlerWar(state, new_ch)
	war_exch, war_q_name, war_key, war_q_type := routing.ExchangePerilTopic, "war", routing.WarRecognitionsPrefix + ".*", pubsub.QueueTypeDurable
	if err := pubsub.SubscribeJSON(conn, war_exch, war_q_name, war_key, war_q_type, war_handler); err != nil {
		log.Fatal("Couldn't subscribe to the war queue:", err)
	}

	for {
		inp := gamelogic.GetInput()

		if len(inp) == 0 {
			continue
		}

		if inp[0] == "spawn" {
			if err := state.CommandSpawn(inp); err != nil {
				fmt.Println(err)
			}
		} else if inp[0] == "move" {
			move, err := state.CommandMove(inp)
			if err != nil {
				fmt.Println(err)
			}

			if err = pubsub.PublishJSON(new_ch, move_exch, move_q_name, move); err != nil {
				fmt.Println("Couldn't publish move message:", err)
			}

			fmt.Println("Moved successfully!")
		} else if inp[0] == "status" {
			state.CommandStatus()
		} else if inp[0] == "help" {
			gamelogic.PrintClientHelp()
		} else if inp[0] == "spam" {
			if len(inp) < 2 {
				fmt.Println("Usage: spam <number>")
				continue
			}

			num, err := strconv.Atoi(inp[1])
			if err != nil {
				fmt.Println("Invalid number input")
				continue
			}

			log_exch, log_key := routing.ExchangePerilTopic, fmt.Sprintf("%v.%v", routing.GameLogSlug, c_name)
			for ; num > 0; num-- {
				log := gamelogic.GetMaliciousLog()
				data := routing.GameLog{CurrentTime: time.Now(), Message: log, Username: c_name}
				if err = pubsub.PublishGob(new_ch, log_exch, log_key, data); err != nil {
					fmt.Println("Couldn't send malicious log:", err)
				}
			}
			fmt.Println("All logs sent successfully!")
		} else if inp[0] == "quit" {
			gamelogic.PrintQuit()
			break
		} else {
			fmt.Println("Unknown command. For help write \"help\"")
		}
	}
}
