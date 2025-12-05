package services

import (
	"fmt"
	"log"
	"time"

	"ecommerce-microservices/inventory-service/domain/entities"
)

type EmailService interface {
    SendOrderConfirmation(order *entities.Order) error
    SendOrderCancellation(order *entities.Order, reason string) error
}

type MockEmailService struct{}

func NewMockEmailService() *MockEmailService {
    return &MockEmailService{}
}

func (s *MockEmailService) SendOrderConfirmation(order *entities.Order) error {
    // Simulate email sending
    log.Printf("📧 [EMAIL SERVICE] Sending confirmation email...")
    
    emailContent := fmt.Sprintf(`
==================================================
                ORDER CONFIRMATION
==================================================
Dear Valued Customer,

Your order has been successfully confirmed!

Order Details:
- Order ID: #%d
- Status: %s  
- Order Date: %s

Your items will be processed and shipped soon.

Thank you for choosing our service!

Best regards,
E-Commerce Team
==================================================
`, order.ID, order.Status, time.Now().Format("2006-01-02 15:04:05"))

    log.Printf("%s", emailContent)
    
    // Simulate network delay
    time.Sleep(50 * time.Millisecond)
    
    log.Printf("✅ [EMAIL SERVICE] Confirmation email sent successfully for Order #%d", order.ID)
    return nil
}

func (s *MockEmailService) SendOrderCancellation(order *entities.Order, reason string) error {
    // Simulate email sending
    log.Printf("📧 [EMAIL SERVICE] Sending cancellation email...")
    
    emailContent := fmt.Sprintf(`
==================================================
                ORDER CANCELLATION
==================================================
Dear Valued Customer,

We regret to inform you that your order has been cancelled.

Order Details:
- Order ID: #%d
- Status: %s
- Cancellation Reason: %s
- Date: %s

We apologize for any inconvenience this may cause.
If you have any questions, please contact our support team.

Best regards,
E-Commerce Team
==================================================
`, order.ID, order.Status, reason, time.Now().Format("2006-01-02 15:04:05"))

    log.Printf("%s", emailContent)
    
    // Simulate network delay
    time.Sleep(50 * time.Millisecond)
    
    log.Printf("❌ [EMAIL SERVICE] Cancellation email sent successfully for Order #%d", order.ID)
    return nil
}