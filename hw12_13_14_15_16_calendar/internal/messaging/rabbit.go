package messaging

import (
	"github.com/streadway/amqp"
)

type RMQ struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	Queue   amqp.Queue
}

func New(url, queueName string) (*RMQ, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, err
	}

	q, err := ch.QueueDeclare(
		queueName, false, false, false, false, nil,
	)
	if err != nil {
		conn.Close()
		ch.Close()
		return nil, err
	}

	return &RMQ{conn: conn, channel: ch, Queue: q}, nil
}

func (r *RMQ) Publish(body []byte) error {
	return r.channel.Publish(
		"", r.Queue.Name, false, false,
		amqp.Publishing{ContentType: "application/json", Body: body},
	)
}

func (r *RMQ) Consume() (<-chan amqp.Delivery, error) {
	return r.channel.Consume(
		r.Queue.Name, "", false, false, false, false, nil,
	)
}

func (r *RMQ) Close() {
	_ = r.channel.Close()
	_ = r.conn.Close()
}
