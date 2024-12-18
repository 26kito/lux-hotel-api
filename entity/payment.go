package entity

type Payment struct {
	ID            uint    `gorm:"primaryKey;autoIncrement"`
	BookingID     uint    `gorm:"not null" json:"booking_id"`
	TotalAmount   float64 `gorm:"type:decimal(10,2);not null" json:"total_amount"`
	PaymentDate   string  `gorm:"type:date;not null" json:"payment_date"`
	Status        string  `gorm:"type:varchar(10);not null" json:"status"`
	PaymentMethod string  `gorm:"type:varchar(20);not null" json:"payment_method"`
}
