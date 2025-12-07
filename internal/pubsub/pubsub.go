package pubsub

import (
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
