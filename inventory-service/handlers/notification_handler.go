package handlers

import (
	"encoding/json"
	"log"

	"ecommerce-microservices/inventory-service/domain/entities"
	"ecommerce-microservices/inventory-service/services"
	"ecommerce-microservices/inventory-service/usecases"

	amqp "github.com/rabbitmq/amqp091-go"
)

type notificationHandler struct {
    orderUsecase usecases.OrderUsecase
    emailService services.EmailService
}

type OrderFailedMessage struct {
    OrderID uint   `json:"order_id"`
    Status  string `json:"status"`
    Reason  string `json:"reason"`
}

func NewNotificationHandler(orderUsecase usecases.OrderUsecase, emailService services.EmailService) *notificationHandler {
    return &notificationHandler{
        orderUsecase: orderUsecase,
        emailService: emailService,
    }
}

func (h *notificationHandler) StartNotificationWorkers() {
    // Start both workers concurrently
    go h.OrderConfirmedWorker()
    go h.OrderFailedWorker()
    
    log.Println("Notification workers started")
    
    // Keep the main goroutine alive
    select {}
}

func (h *notificationHandler) OrderConfirmedWorker() {
    conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
    if err != nil {
        log.Fatalf("Failed to connect to RabbitMQ for order_confirmed: %v", err)
    }
    defer conn.Close()

    ch, err := conn.Channel()
    if err != nil {
        log.Fatalf("Failed to open channel for order_confirmed: %v", err)
    }
    defer ch.Close()

    // Declare the queue
    queue, err := ch.QueueDeclare(
        "order_confirmed", // queue name
        true,             // durable
        false,            // delete when unused
        false,            // exclusive
        false,            // no-wait
        nil,              // arguments
    )
    if err != nil {
        log.Fatalf("Failed to declare order_confirmed queue: %v", err)
    }

    // Set QoS
    err = ch.Qos(1, 0, false)
    if err != nil {
        log.Fatalf("Failed to set QoS for order_confirmed: %v", err)
    }

    // Start consuming
    msgs, err := ch.Consume(
        queue.Name,
        "notification_service_confirmed", // consumer tag
        false, // auto-ack
        false, // exclusive
        false, // no-local
        false, // no-wait
        nil,   // args
    )
    if err != nil {
        log.Fatalf("Failed to register consumer for order_confirmed: %v", err)
    }

    log.Printf("Notification worker started for order_confirmed messages...")

    for msg := range msgs {
        h.processOrderConfirmed(msg)
    }
}

func (h *notificationHandler) OrderFailedWorker() {
    conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
    if err != nil {
        log.Fatalf("Failed to connect to RabbitMQ for order_failed: %v", err)
    }
    defer conn.Close()

    ch, err := conn.Channel()
    if err != nil {
        log.Fatalf("Failed to open channel for order_failed: %v", err)
    }
    defer ch.Close()

    // Declare the queue
    queue, err := ch.QueueDeclare(
        "order_failed", // queue name
        true,          // durable
        false,         // delete when unused
        false,         // exclusive
        false,         // no-wait
        nil,           // arguments
    )
    if err != nil {
        log.Fatalf("Failed to declare order_failed queue: %v", err)
    }

    // Set QoS
    err = ch.Qos(1, 0, false)
    if err != nil {
        log.Fatalf("Failed to set QoS for order_failed: %v", err)
    }

    // Start consuming
    msgs, err := ch.Consume(
        queue.Name,
        "notification_service_failed", // consumer tag
        false, // auto-ack
        false, // exclusive
        false, // no-local
        false, // no-wait
        nil,   // args
    )
    if err != nil {
        log.Fatalf("Failed to register consumer for order_failed: %v", err)
    }

    log.Printf("Notification worker started for order_failed messages...")

    for msg := range msgs {
        h.processOrderFailed(msg)
    }
}

func (h *notificationHandler) processOrderConfirmed(msg amqp.Delivery) {
    log.Printf("Received order_confirmed message: %s", msg.Body)

    var orderMsg OrderConfirmedMessage
    if err := json.Unmarshal(msg.Body, &orderMsg); err != nil {
        log.Printf("Failed to parse order_confirmed message: %v", err)
        msg.Nack(false, false)
        return
    }

    // Get order details for email
    order, err := h.orderUsecase.GetOrder(orderMsg.OrderID)
    if err != nil {
        log.Printf("Failed to get order %d details: %v", orderMsg.OrderID, err)
        msg.Nack(false, true) // Retry this one
        return
    }

    // Send confirmation email
    err = h.sendOrderConfirmationEmail(order)
    if err != nil {
        log.Printf("Failed to send confirmation email for order %d: %v", orderMsg.OrderID, err)
        msg.Nack(false, true) // Retry sending email
        return
    }

    msg.Ack(false)
    log.Printf("Successfully sent confirmation email for order %d", orderMsg.OrderID)
}

func (h *notificationHandler) processOrderFailed(msg amqp.Delivery) {
    log.Printf("Received order_failed message: %s", msg.Body)

    var orderMsg OrderFailedMessage
    if err := json.Unmarshal(msg.Body, &orderMsg); err != nil {
        log.Printf("Failed to parse order_failed message: %v", err)
        msg.Nack(false, false)
        return
    }

    // Get order details for email
    order, err := h.orderUsecase.GetOrder(orderMsg.OrderID)
    if err != nil {
        log.Printf("Failed to get order %d details: %v", orderMsg.OrderID, err)
        msg.Nack(false, true) // Retry this one
        return
    }

    // Send cancellation email
    err = h.sendOrderCancellationEmail(order, orderMsg.Reason)
    if err != nil {
        log.Printf("Failed to send cancellation email for order %d: %v", orderMsg.OrderID, err)
        msg.Nack(false, true) // Retry sending email
        return
    }

    msg.Ack(false)
    log.Printf("Successfully sent cancellation email for order %d", orderMsg.OrderID)
}

// Simulate sending confirmation email
func (h *notificationHandler) sendOrderConfirmationEmail(order *entities.Order) error {
    return h.emailService.SendOrderConfirmation(order)
}

// Simulate sending cancellation email
func (h *notificationHandler) sendOrderCancellationEmail(order *entities.Order, reason string) error {
    return h.emailService.SendOrderCancellation(order, reason)
}