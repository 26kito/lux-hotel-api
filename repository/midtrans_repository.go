package repository

import (
	"lux-hotel/entity"

	"gorm.io/gorm"
)

type MidtransRepository interface {
	HandleTopUpCallback(entity.MidtransCallbackResponse)
	HandleBookingCallback(entity.MidtransCallbackResponse)
}

type midtransRepository struct {
	DB *gorm.DB
}

func NewMidtransRepository(db *gorm.DB) MidtransRepository {
	return &midtransRepository{DB: db}
}

func (mr *midtransRepository) HandleTopUpCallback(payload entity.MidtransCallbackResponse) {
	if payload.TransactionStatus == "settlement" {
		var user entity.User
		var transaction entity.TopUpTransaction
		var payment entity.Payment

		mr.DB.Where("order_id = ?", payload.OrderID).First(&transaction)
		mr.DB.Model(&user).Where("user_id = ?", transaction.UserID)

		newBalance := user.Balance + transaction.Amount
		mr.DB.Model(&user).Update("balance", newBalance)

		mr.DB.Model(&transaction).Where("order_id = ?", payload.OrderID).Update("transaction_status", "settlement")
		mr.DB.Model(&payment).Where("order_id = ?", payload.OrderID).Updates(map[string]interface{}{
			"payment_status": "settlement",
			"payment_date":   payload.TransactionTime,
		})
	}

	if payload.TransactionStatus == "expire" || payload.TransactionStatus == "cancel" {
		var transaction entity.TopUpTransaction
		var payment entity.Payment

		mr.DB.Model(&transaction).Where("order_id = ?", payload.OrderID).Update("transaction_status", payload.TransactionStatus)
		mr.DB.Model(&payment).Where("order_id = ?", payload.OrderID).Update("payment_status", payload.TransactionStatus)
	}
}

func (mr *midtransRepository) HandleBookingCallback(payload entity.MidtransCallbackResponse) {
	if payload.TransactionStatus == "settlement" {
		var booking entity.Booking
		var payment entity.Payment
		var room entity.Room

		mr.DB.Where("order_id = ?", payload.OrderID).First(&booking)
		mr.DB.Model(&room).Where("hotel_id = ? AND id = ?", booking.HotelID, booking.RoomID).Update("status", "occupied")
		mr.DB.Model(&booking).Update("booking_status", "settlement")
		mr.DB.Model(&payment).Where("order_id = ?", payload.OrderID).Updates(map[string]interface{}{
			"payment_status": "settlement",
			"payment_date":   payload.TransactionTime,
		})
	}

	if payload.TransactionStatus == "expire" || payload.TransactionStatus == "cancel" {
		var booking entity.Booking
		var payment entity.Payment

		mr.DB.Model(&booking).Where("order_id = ?", payload.OrderID).Update("booking_status", payload.TransactionStatus)
		mr.DB.Model(&payment).Where("order_id = ?", payload.OrderID).Update("payment_status", payload.TransactionStatus)
	}
}
