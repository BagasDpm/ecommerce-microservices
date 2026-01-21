package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"ecommerce-microservices/inventory-service/domain/entities"
	"ecommerce-microservices/inventory-service/infrastructure/messaging"
	"ecommerce-microservices/inventory-service/usecases"

	amqp "github.com/rabbitmq/amqp091-go"
)

type inventoryHandler struct {
	orderUsecase     usecases.OrderUsecase
	OrderItemUsecase usecases.OrderItemUsecase
	ProductUsecase   usecases.ProductUsecase
	publisher        *messaging.RabbitMQPublisher
}

type OrderPlacedMessage struct {
	OrderID uint `json:"order_id"`
}

type OrderConfirmedMessage struct {
	OrderID uint   `json:"order_id"`
	Status  string `json:"status"`
}

func NewInventoryHandler(orderUsecase usecases.OrderUsecase, orderItemUsecase usecases.OrderItemUsecase, productUsecase usecases.ProductUsecase, publisher *messaging.RabbitMQPublisher) *inventoryHandler {
	return &inventoryHandler{
		orderUsecase:     orderUsecase,
		OrderItemUsecase: orderItemUsecase,
		ProductUsecase:   productUsecase,
		publisher:        publisher,
	}
}

func (h *inventoryHandler) OrderPlacedWorker() {
	// Set up RabbitMQ connection and channel
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Failed to open a channel: %v", err)
	}
	defer ch.Close()

	// Declare the queue
	queue, err := ch.QueueDeclare(
		"order_placed", // queue name
		true,          // durable
		false,         // delete when unused
		false,         // exclusive
		false,         // no-wait
		nil,           // arguments
	)
	if err != nil {
		log.Fatalf("Failed to declare queue: %v", err)
	}

	// Set QoS to process one message at a time
	err = ch.Qos(1, 0, false)
	if err != nil {
		log.Fatalf("Failed to set QoS: %v", err)
	}

	// Start consuming messages
	msgs, err := ch.Consume(
		queue.Name, // queue
	"",         // consumer
	 false,      // auto-ack (set to false for manual ack)
	 false,      // exclusive
	 false,      // no-local
	 false,      // no-wait
	 nil,        // args
	)
	if err != nil {
		log.Fatalf("Failed to register consumer: %v", err)
	}

	log.Printf("Inventory worker started. Waiting for order_placed messages...")

	// Process messages
	for msg := range msgs {
		h.processOrderPlaced(msg)
	}
}

func (h *inventoryHandler) processOrderPlaced(msg amqp.Delivery) {
	log.Printf("Received order_placed message: %s", msg.Body)

	// Parse the message
	var orderMsg OrderPlacedMessage
	if err := json.Unmarshal(msg.Body, &orderMsg); err != nil {
		log.Printf("Failed to parse message: %v", err)
		msg.Nack(false, false) // Don't requeue invalid messages
		return
	}

	// Process the order with transaction
	err := h.processInventoryCheck(orderMsg.OrderID)
	if err != nil {
		log.Printf("Failed to process order %d: %v", orderMsg.OrderID, err)
		// Don't requeue, let it go to dead letter queue or handle retry logic
		msg.Nack(false, false)
		return
	}

	// Acknowledge the message after successful processing
	msg.Ack(false)
	log.Printf("Successfully processed order %d", orderMsg.OrderID)
}

func (h *inventoryHandler) processInventoryCheck(orderID uint) error {
    ctx := context.Background()
    
    // Get order details
    _, err := h.orderUsecase.GetOrder(orderID)
    if err != nil {
        return fmt.Errorf("failed to get order: %w", err)
    }

    // Get order items
    orderItems, err := h.OrderItemUsecase.GetByOrderID(orderID)
    if err != nil {
        return fmt.Errorf("failed to get order items: %w", err)
    }

    if len(orderItems) == 0 {
        return h.handleOrderFailure(orderID, "no items in order")
    }

    // Prepare stock updates
    stockUpdates := make([]struct {
        ProductID uint
        Quantity  int
    }, 0, len(orderItems))

    // Validate all products have sufficient stock using cache
    for _, item := range orderItems {
        product, err := h.ProductUsecase.GetByIDWithCache(ctx, item.ProductID)
        if err != nil {
            return h.handleOrderFailure(orderID, fmt.Sprintf("product %d not found", item.ProductID))
        }

        if product.Stock < item.Quantity {
            return h.handleOrderFailure(orderID,
                fmt.Sprintf("insufficient stock for product %d: available=%d, required=%d",
                    item.ProductID, product.Stock, item.Quantity))
        }

        stockUpdates = append(stockUpdates, struct {
            ProductID uint
            Quantity  int
        }{
            ProductID: item.ProductID,
            Quantity:  item.Quantity,
        })
    }

    // All validations passed, now update inventory atomically
    err = h.ProductUsecase.DecreaseStock(stockUpdates)
    if err != nil {
        return h.handleOrderFailure(orderID, fmt.Sprintf("failed to update inventory: %v", err))
    }

    // Update order status to CONFIRMED
    err = h.orderUsecase.UpdateStatus(orderID, entities.OrderStatusConfirmed)
    if err != nil {
        log.Printf("Warning: Failed to update order status to confirmed for order %d: %v", orderID, err)
        // Continue anyway since inventory was already decremented
    }

    // Publish order_confirmed event
    return h.publishOrderConfirmed(orderID)
}

func (h *inventoryHandler) handleOrderFailure(orderID uint, reason string) error {
	log.Printf("Order %d failed: %s", orderID, reason)

	// Update order status to CANCELLED
	err := h.orderUsecase.UpdateStatus(orderID, entities.OrderStatusCancelled)
	if err != nil {
		log.Printf("Failed to update order status to cancelled for order %d: %v", orderID, err)
	}

	// Publish order_failed event
	return h.publishOrderFailed(orderID, reason)
}

func (h *inventoryHandler) publishOrderConfirmed(orderID uint) error {
	message := OrderConfirmedMessage{
		OrderID: orderID,
		Status:  string(entities.OrderStatusConfirmed),
	}

	messageBytes, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal order_confirmed message: %w", err)
	}

	err = h.publisher.Publish("order_confirmed", messageBytes)
	if err != nil {
		return fmt.Errorf("failed to publish order_confirmed event: %w", err)
	}

	log.Printf("Published order_confirmed event for order %d", orderID)
	return nil
}

func (h *inventoryHandler) publishOrderFailed(orderID uint, reason string) error {
	message := map[string]interface{}{
		"order_id": orderID,
		"status":   string(entities.OrderStatusCancelled),
		"reason":   reason,
	}

	messageBytes, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal order_failed message: %w", err)
	}

	err = h.publisher.Publish("order_failed", messageBytes)
	if err != nil {
		return fmt.Errorf("failed to publish order_failed event: %w", err)
	}

	log.Printf("Published order_failed event for order %d: %s", orderID, reason)
	return nil
}