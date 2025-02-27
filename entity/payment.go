package entity

import "time"

type Payment struct {
	ID              uint       `gorm:"primaryKey;autoIncrement"`
	PaymentID       string     `gorm:"unique;not null" json:"payment_id"`
	OrderID         string     `gorm:"unique;not null" json:"order_id"`
	UserID          uint       `gorm:"not null" json:"user_id"`
	TotalAmount     float64    `gorm:"type:decimal(10,2);not null" json:"total_amount"`
	TransactionType string     `gorm:"type:varchar(20);not null" json:"transaction_type"` // "topup" or "booking"
	PaymentDate     *time.Time `gorm:"type:date" json:"payment_date"`
	PaymentStatus   string     `gorm:"type:varchar(10);not null" json:"payment_status"`
	PaymentMethod   string     `gorm:"type:varchar(20);not null" json:"payment_method"`
	CreatedAt       time.Time  `gorm:"type:timestamp" json:"created_at"`
	UpdatedAt       time.Time  `gorm:"type:timestamp" json:"updated_at"`
}

type PaymentPayload struct {
	OrderID       string `json:"order_id"`
	PaymentMethod string `json:"payment_method"`
}

type PaymentResponse struct {
	TransactionID     string  `json:"transaction_id"`
	TransactionStatus string  `json:"transaction_status"`
	Amount            float64 `json:"amount"`
	PaymentType       string  `json:"payment_type"`
	Bank              string  `json:"bank,omitempty"`
	VANumber          string  `json:"va_number,omitempty"`
}
