package pubsub

import (
	"log"
	"bytes"
	"context"
	"encoding/gob"
	"encoding/json"
	amqp "github.com/rabbitmq/amqp091-go"
)

type AckType int

const (
	Ack AckType = iota
	NackRequeue
	NackDiscard
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

	args := amqp.Table{"x-dead-letter-exchange": "peril_dlx"}

	new_queue, err := new_ch.QueueDeclare(queueName, queue.durable, queue.autoDelete, queue.exclusive, false, args)
	if err != nil {
		return nil, amqp.Queue{}, err
	}

	if err = new_ch.QueueBind(queueName, key, exchange, false, nil); err != nil {
		return nil, amqp.Queue{}, err
	}

	return new_ch, new_queue, nil
}

func SubscribeJSON[T any](conn *amqp.Connection, exchange, queueName, key string, queueType SimpleQueueType, handler func(T) AckType) error {
	unmarshaller := func(in []byte) (T, error) {
		var unm_val T
		body_reader := bytes.NewReader(in)
		decoder := json.NewDecoder(body_reader)
		if err := decoder.Decode(&unm_val); err != nil {
			return unm_val, err
		}

		return unm_val, nil
	}

	if err := subscribe(conn, exchange, queueName, key, queueType, handler, unmarshaller); err != nil {
		return err
	}

	return nil
}

func PublishGob[T any](ch *amqp.Channel, exchange, key string, val T) error {
	var res bytes.Buffer
	gob_enc := gob.NewEncoder(&res)
	if err := gob_enc.Encode(val); err != nil {
		return err
	}

	msg := amqp.Publishing{
		ContentType: "application/gob",
		Body:        res.Bytes(),
	}

	if err := ch.PublishWithContext(context.Background(), exchange, key, false, false, msg); err != nil {
		return err
	}

	return nil
}

func SubscribeGob[T any](conn *amqp.Connection, exchange, queueName, key string, queueType SimpleQueueType, handler func(T) AckType) error {
	unmarshaller := func(in []byte) (T, error) {
		var unm_val T
		gob_dec := gob.NewDecoder(bytes.NewReader(in))
		if err := gob_dec.Decode(&unm_val); err != nil {
			return unm_val, err
		}
		return unm_val, nil
	}

	if err := subscribe(conn, exchange, queueName, key, queueType, handler, unmarshaller); err != nil {
		return err
	}

	return nil
}

func subscribe[T any](conn *amqp.Connection, exchange, queueName, key string, simpleQueueType SimpleQueueType, handler func(T) AckType, unmarshaller func([]byte) (T, error)) error {
	new_ch, _, err := DeclareAndBind(conn, exchange, queueName, key, simpleQueueType)
	if err != nil {
		return err
	}

	if err = new_ch.Qos(10, 0, false); err != nil {
		return err
	}

	delivery, err := new_ch.Consume(queueName, "", false, false, false, false, nil)
	if err != nil {
		return err
	}

	go func() {
		for elem := range delivery {
			val, err := unmarshaller(elem.Body)
			if err != nil {
				log.Println("Couldn't decode message:", err)
				continue
			}

			ack_type := handler(val)

			switch ack_type {
			case Ack:
				if err := elem.Ack(false); err != nil {
					log.Println("Couldn't acknowledge delivery message:", err)
				}
			case NackRequeue:
				if err := elem.Nack(false, true); err != nil {
					log.Println("Couldn't not acknowledge delivery message (requeue:", err)
				}
			case NackDiscard:
				if err := elem.Nack(false, false); err != nil {
					log.Println("Couldn't not acknowledge delivery message (discard):", err)
				}
			}
		}
	}()

	return nil
}
