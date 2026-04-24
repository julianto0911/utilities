package utilities

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"sync"
	"sync/atomic"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/mock"
)

// RabbitMQ connection tuning.
//
// The Heartbeat value is the client's proposal; the broker may negotiate it
// down. Going lower than the amqp091-go default (10s) does not help because
// the broker side usually caps it. 15s is a safe middle ground that still
// allows the broker to detect dead peers within ~30s.
const (
	defaultRabbitMQHeartbeat   = 15 * time.Second
	defaultRabbitMQDialTimeout = 10 * time.Second
	defaultRabbitMQKeepAlive   = 30 * time.Second
)

const (
	consumeInitialBackoff  = time.Second
	consumeMaxBackoff      = 30 * time.Second
	consumeDrainRetryPause = 200 * time.Millisecond
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

var errRabbitMQClosed = errors.New("rabbitmq: client closed")

func NewRabbitMQ(user, password, host, port string) (RabbitMQ, error) {
	uri := rabbitMQString(user, password, host, port)
	r := &rabbitMQ{uri: uri}
	r.mu.Lock()
	defer r.mu.Unlock()
	if err := r.dialFreshLocked(); err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}
	log.Printf("rabbitmq: connected (heartbeat=%s)", defaultRabbitMQHeartbeat)
	return r, nil
}

type RabbitMQ interface {
	Publish(queueName string, message []byte) error
	Consume(queueName string, handler func([]byte) error) error
	Close()
}

type rabbitMQ struct {
	mu      sync.Mutex
	uri     string
	conn    *amqp.Connection
	channel *amqp.Channel
	closed  atomic.Bool
}

func (r *rabbitMQ) dialFreshLocked() error {
	cfg := amqp.Config{
		Heartbeat: defaultRabbitMQHeartbeat,
		Locale:    "en_US",
		Dial: func(network, addr string) (net.Conn, error) {
			d := net.Dialer{
				Timeout:   defaultRabbitMQDialTimeout,
				KeepAlive: defaultRabbitMQKeepAlive,
			}
			return d.Dial(network, addr)
		},
		Properties: amqp.Table{
			"connection_name": rabbitMQConnectionName(),
			"product":         "xs-utilities",
		},
	}
	conn, err := amqp.DialConfig(r.uri, cfg)
	if err != nil {
		return err
	}
	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return err
	}
	r.conn = conn
	r.channel = ch
	return nil
}

// rabbitMQConnectionName returns a label shown in `rabbitmqctl list_connections`
// and the management UI so operators can identify which service owns a
// connection. It prefers APP_NAME, falls back to hostname, then "xs-service".
func rabbitMQConnectionName() string {
	if name := os.Getenv("APP_NAME"); name != "" {
		return name
	}
	if host, err := os.Hostname(); err == nil && host != "" {
		return host
	}
	return "xs-service"
}

func (r *rabbitMQ) teardownLocked() {
	if r.channel != nil {
		_ = r.channel.Close()
		r.channel = nil
	}
	if r.conn != nil {
		_ = r.conn.Close()
		r.conn = nil
	}
}

// Publish sends a message to a queue
func (r *rabbitMQ) Publish(queueName string, message []byte) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.closed.Load() {
		return errRabbitMQClosed
	}
	err := r.publishLocked(queueName, message)
	if err != nil {
		log.Printf("rabbitmq: publish failed, reconnecting: %v", err)
		r.teardownLocked()
		if err2 := r.dialFreshLocked(); err2 != nil {
			return fmt.Errorf("publish: %w; reconnect: %v", err, err2)
		}
		err = r.publishLocked(queueName, message)
		if err != nil {
			return fmt.Errorf("failed to publish message after reconnect: %w", err)
		}
	}
	return nil
}

func (r *rabbitMQ) publishLocked(queueName string, message []byte) error {
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

// Consume starts consuming messages from a queue. It runs a background loop
// that re-declares the consumer after connection/channel loss.
func (r *rabbitMQ) Consume(queueName string, handler func([]byte) error) error {
	if r.closed.Load() {
		return errRabbitMQClosed
	}
	go r.consumeLoop(queueName, handler)
	return nil
}

func (r *rabbitMQ) consumeLoop(queueName string, handler func([]byte) error) {
	backoff := consumeInitialBackoff
	for !r.closed.Load() {
		msgs, err := r.declareAndConsumeLocked(queueName)
		if err != nil {
			log.Printf("rabbitmq: consume setup failed for queue %q: %v", queueName, err)
			if stopped := r.sleepBackoff(&backoff); stopped {
				return
			}
			r.mu.Lock()
			if !r.closed.Load() {
				r.teardownLocked()
				if dialErr := r.dialFreshLocked(); dialErr != nil {
					log.Printf("rabbitmq: reconnect after setup error failed: %v", dialErr)
				} else {
					log.Printf("rabbitmq: reconnected after consume setup error (queue=%q)", queueName)
				}
			}
			r.mu.Unlock()
			continue
		}
		backoff = consumeInitialBackoff

		for msg := range msgs {
			if r.closed.Load() {
				return
			}
			if err := handler(msg.Body); err != nil {
				log.Printf("rabbitmq: handler error: %v", err)
			}
		}

		if r.closed.Load() {
			return
		}
		log.Printf("rabbitmq: delivery channel closed for queue %q, reconnecting", queueName)

		r.mu.Lock()
		if !r.closed.Load() {
			r.teardownLocked()
			if dialErr := r.dialFreshLocked(); dialErr != nil {
				log.Printf("rabbitmq: reconnect after delivery end failed: %v", dialErr)
			} else {
				log.Printf("rabbitmq: reconnected after delivery end (queue=%q)", queueName)
			}
		}
		r.mu.Unlock()

		time.Sleep(consumeDrainRetryPause)
		if r.closed.Load() {
			return
		}
	}
}

func (r *rabbitMQ) declareAndConsumeLocked(queueName string) (<-chan amqp.Delivery, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.closed.Load() {
		return nil, errRabbitMQClosed
	}
	if r.channel == nil || r.conn == nil {
		return nil, errors.New("rabbitmq: not connected")
	}

	q, err := r.channel.QueueDeclare(
		queueName,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to declare queue: %w", err)
	}

	msgs, err := r.channel.Consume(
		q.Name,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to register consumer: %w", err)
	}
	return msgs, nil
}

func (r *rabbitMQ) sleepBackoff(backoff *time.Duration) (stop bool) {
	time.Sleep(*backoff)
	if *backoff < consumeMaxBackoff {
		*backoff *= 2
		if *backoff > consumeMaxBackoff {
			*backoff = consumeMaxBackoff
		}
	}
	return r.closed.Load()
}

// Close closes the RabbitMQ connection
func (r *rabbitMQ) Close() {
	r.closed.Store(true)
	r.mu.Lock()
	defer r.mu.Unlock()
	r.teardownLocked()
	log.Printf("rabbitmq: client closed")
}
