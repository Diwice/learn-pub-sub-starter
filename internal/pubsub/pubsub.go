package pubsub

import (
	"context"
	"encoding/json"
	amqp "github.com/rabbitmq/amqp091-go"
)

func PublishJSON[T any](ch *amqp.Channel, exchange, key string, val T) error {
	b_val, err := json.Marshal(val)
	if err != nil {
		return err
	}

	msg := amqp.Publishing{
		//DeliveryMode: amqp.Persistent,
		//Timestamp:    time.Now(),
		ContentType:  "application/json",
		Body:         b_val,
	}

	if err = ch.PublishWithContext(context.Background(), exchange, key, false, false, msg); err != nil {
		return err
	}

	return nil
}
