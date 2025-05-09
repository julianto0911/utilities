package utilities

import (
	"context"
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/mock"
)

func GetRabbitMQConfig() RabbitMQConfig {
	return RabbitMQConfig{
		Host:     EnvString("RABBIT_HOST"),
		Port:     EnvString("RABBIT_PORT"),
		User:     EnvString("RABBIT_USER"),
		Password: EnvString("RABBIT_PASSWORD"),
	}
}

type RabbitMQConfig struct {
	Host     string
	Port     string
	User     string
	Password string
}

type MockRabbitMQ struct {
	mock.Mock
}

func (m *MockRabbitMQ) Publish(queueName string, message []byte) error {
	args := m.Called(queueName, message)
	return args.Error(0)
}

func (m *MockRabbitMQ) Consume(queueName string, handler func([]byte) error) error {
	args := m.Called(queueName, handler)
	return args.Error(0)
}

func (m *MockRabbitMQ) Close() {
	m.Called()
}

func rabbitMQString(user, password, host, port string) string {
	return fmt.Sprintf("amqp://%s:%s@%s:%s", user, password, host, port)
}

func NewRabbitMQ(user, password, host, port string) (RabbitMQ, error) {
	uri := rabbitMQString(user, password, host, port)
	conn, err := amqp.Dial(uri)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	return &rabbitMQ{
		conn:    conn,
		channel: ch,
	}, nil
}

type RabbitMQ interface {
	Publish(queueName string, message []byte) error
	Consume(queueName string, handler func([]byte) error) error
	Close()
}

type rabbitMQ struct {
	conn    *amqp.Connection
	channel *amqp.Channel
}

// Publish sends a message to a queue
func (r *rabbitMQ) Publish(queueName string, message []byte) error {
	q, err := r.channel.QueueDeclare(
		queueName, // name
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare queue: %w", err)
	}

	err = r.channel.PublishWithContext(
		context.Background(),
		"",     // exchange
		q.Name, // routing key
		false,  // mandatory
		false,  // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        message,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	return nil
}

// Consume starts consuming messages from a queue
func (r *rabbitMQ) Consume(queueName string, handler func([]byte) error) error {
	q, err := r.channel.QueueDeclare(
		queueName, // name
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare queue: %w", err)
	}

	msgs, err := r.channel.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		return fmt.Errorf("failed to register consumer: %w", err)
	}

	go func() {
		for msg := range msgs {
			if err := handler(msg.Body); err != nil {
				log.Printf("Error handling message: %v", err)
			}
		}
	}()

	return nil
}

// Close closes the RabbitMQ connection
func (r *rabbitMQ) Close() {
	if r.channel != nil {
		r.channel.Close()
	}
	if r.conn != nil {
		r.conn.Close()
	}
}
