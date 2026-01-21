package messaging

import (
	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQPublisher struct {
    conn    *amqp.Connection
    channel *amqp.Channel
}

func NewRabbitMQPublisher(url string) (*RabbitMQPublisher, error) {
    conn, err := amqp.Dial(url)
    if err != nil {
        return nil, err
    }

    ch, err := conn.Channel()
    if err != nil {
        conn.Close()
        return nil, err
    }

    return &RabbitMQPublisher{
        conn:    conn,
        channel: ch,
    }, nil
}

func (p *RabbitMQPublisher) Publish(queueName string, message []byte) error {
    // Declare queue
    _, err := p.channel.QueueDeclare(
        queueName,
        true,  // durable
        false, // delete when unused
        false, // exclusive
        false, // no-wait
        nil,   // arguments
    )
    if err != nil {
        return err
    }

    // Publish message
    return p.channel.Publish(
        "",        // exchange
        queueName, // routing key
        false,     // mandatory
        false,     // immediate
        amqp.Publishing{
            ContentType: "application/json",
            Body:        message,
        })
}

func (p *RabbitMQPublisher) Close() {
    if p.channel != nil {
        p.channel.Close()
    }
    if p.conn != nil {
        p.conn.Close()
    }
}