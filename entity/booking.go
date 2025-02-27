package entity

import "time"

type Booking struct {
	ID            uint      `gorm:"primaryKey;autoIncrement"`
	OrderID       string    `gorm:"unique;not null" json:"order_id"`
	BookingCode   string    `gorm:"type:varchar(10);not null" json:"booking_code"`
	GuestID       uint      `gorm:"not null" json:"guest_id"`
	HotelID       uint      `gorm:"not null" json:"hotel_id"`
	RoomID        uint      `gorm:"not null" json:"room_id"`
	CheckIn       string    `gorm:"type:date;not null" json:"check_in"`
	CheckOut      string    `gorm:"type:date;not null" json:"check_out"`
	TotalDays     int       `gorm:"not null" json:"total_days"`
	TotalPrice    float64   `gorm:"type:decimal(10,2);not null" json:"total_price"`
	BookingStatus string    `gorm:"type:varchar(10);default:pending" json:"booking_status"`
	CreatedAt     time.Time `gorm:"type:timestamp" json:"created_at"`
	UpdatedAt     time.Time `gorm:"type:timestamp" json:"updated_at"`
}

type BookingRequest struct {
	RoomID   uint   `json:"room_id"`
	CheckIn  string `json:"check_in"`
	CheckOut string `json:"check_out"`
}

type BookingHistoryResponse struct {
	OrderID       string  `json:"order_id"`
	BookingCode   string  `json:"booking_code"`
	HotelName     string  `json:"hotel_name"`
	RoomNumber    string  `json:"room_number"`
	CheckIn       string  `json:"check_in"`
	CheckOut      string  `json:"check_out"`
	TotalDays     int     `json:"total_days"`
	TotalPrice    float64 `json:"total_price"`
	BookingDate   string  `json:"booking_date"`
	BookingStatus string  `json:"booking_status"`
}
