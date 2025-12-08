package pubsub

import (
	"log"
	"bytes"
	"context"
	"encoding/json"
	amqp "github.com/rabbitmq/amqp091-go"
)

type SimpleQueueType struct{
	durable    bool
	autoDelete bool
	exclusive  bool
}

var (
	QueueTypeDurable = SimpleQueueType{durable: true, autoDelete: false, exclusive: false}
	QueueTypeTransient = SimpleQueueType{durable: false, autoDelete: true, exclusive: true}
)

func PublishJSON[T any](ch *amqp.Channel, exchange, key string, val T) error {
	b_val, err := json.Marshal(val)
	if err != nil {
		return err
	}

	msg := amqp.Publishing{
		ContentType:  "application/json",
		Body:         b_val,
	}

	if err = ch.PublishWithContext(context.Background(), exchange, key, false, false, msg); err != nil {
		return err
	}

	return nil
}

func DeclareAndBind(conn *amqp.Connection, exchange, queueName, key string, queue SimpleQueueType) (*amqp.Channel, amqp.Queue, error) {
	new_ch, err := conn.Channel()
	if err != nil {
		return nil, amqp.Queue{}, err
	}

	new_queue, err := new_ch.QueueDeclare(queueName, queue.durable, queue.autoDelete, queue.exclusive, false, nil)
	if err != nil {
		return nil, amqp.Queue{}, err
	}

	if err = new_ch.QueueBind(queueName, key, exchange, false, nil); err != nil {
		return nil, amqp.Queue{}, err
	}

	return new_ch, new_queue, nil
}

func SubscribeJSON[T any](conn *amqp.Connection, exchange, queueName, key string, queueType SimpleQueueType, handler func(T)) error {
	new_ch, _, err := DeclareAndBind(conn, exchange, queueName, key, queueType)
	if err != nil {
		return err
	}

	delivery, err := new_ch.Consume(queueName, "", false, false, false, false, nil)
	if err != nil {
		return err
	}

	go func() {
		for elem := range delivery {
			var unm_elem T
			body_reader := bytes.NewReader(elem.Body)
			decoder := json.NewDecoder(body_reader)
			if err := decoder.Decode(&unm_elem); err != nil {
				log.Println("Couldn't decode message:", err)
				continue
			}

			handler(unm_elem)

			if err := elem.Ack(false); err != nil {
				log.Println("Couldn't acknowledge delivery message:", err)
			}
		}
	}()

	return nil
}
