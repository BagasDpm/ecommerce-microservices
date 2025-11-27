package messaging

import (
	"ecommerce-microservices/config"
	"encoding/json"
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQPublisher struct {
	conn    *amqp.Connection
	channel *amqp.Channel
}

type OrderPlacedMessage struct {
	OrderID uint `json:"order_id"`
}

func NewRabbitMQPublisher(cfg *config.RabbitMQConfig) (*RabbitMQPublisher, error) {
	url := fmt.Sprintf("amqp://%s:%s@%s:%s/",
		cfg.User, cfg.Password, cfg.Host, cfg.Port)

	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	channel, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	queues := []string{"order_placed", "order_confirmed", "order_failed"}
	for _, queueName := range queues {
		_, err = channel.QueueDeclare(
			queueName,
			true,
			false,
			false,
			false,
			nil,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to declare queue %s: %w", queueName, err)
		}
	}

	log.Println("RabbitMQ connected successfully")
	return &RabbitMQPublisher{
		conn:    conn,
		channel: channel,
	}, nil
}

func (r *RabbitMQPublisher) PublishOrderPlaced(orderID uint) error {
	message := OrderPlacedMessage{OrderID: orderID}
	body, err := json.Marshal(message)
	if err != nil {
		return err
	}

	return r.channel.Publish(
		"",
		"order_placed",
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
}

func (r *RabbitMQPublisher) Close() {
	if r.channel != nil {
		r.channel.Close()
	}
	if r.conn != nil {
		r.conn.Close()
	}
}
