package queue

import (
	"encoding/json"
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQConnection struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	queue   string
}

type RabbitMQ interface {
	Publish(data interface{}) error
	Consume() (<-chan amqp.Delivery, error)
	Close()
}

func NewConnection(uri, queue string) (*RabbitMQConnection, error) {
	conn, err := amqp.Dial(uri)
	if err != nil {
		return nil, err
	}

	channel, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, err
	}

	_, err = channel.QueueDeclare(
		queue,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		channel.Close()
		conn.Close()
		return nil, err
	}

	return &RabbitMQConnection{
		conn:    conn,
		channel: channel,
		queue:   queue,
	}, nil
}

func (p *RabbitMQConnection) Publish(data interface{}) error {
	body, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to encode data: %w", err)
	}

	return p.channel.Publish(
		"",
		p.queue,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
}

func (p *RabbitMQConnection) Consume() (<-chan amqp.Delivery, error) {
	msgs, err := p.channel.Consume(
		p.queue, // queue name
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to consume queue: %w", err)
	}

	return msgs, nil
}

func (p *RabbitMQConnection) Close() {
	p.channel.Close()
	p.conn.Close()
}
