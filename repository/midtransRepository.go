package repository

import (
	"lux-hotel/entity"

	"gorm.io/gorm"
)

type MidtransRepository interface {
	HandleTopUpCallback(entity.MidtransCallbackResponse)
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
		mr.DB.Model(&user).Where("user_id = ?", transaction.UserID).Update("balance", user.Balance+transaction.Amount)
		mr.DB.Model(&transaction).Where("order_id = ?", payload.OrderID).Update("transaction_status", "settlement")
		mr.DB.Model(&payment).Where("order_id = ?", payload.OrderID).Update("payment_status", "settlement")
	}

	if payload.TransactionStatus == "expire" || payload.TransactionStatus == "cancel" {
		var transaction entity.TopUpTransaction
		var payment entity.Payment

		mr.DB.Model(&transaction).Where("order_id = ?", payload.OrderID).Update("transaction_status", payload.TransactionStatus)
		mr.DB.Model(&payment).Where("order_id = ?", payload.OrderID).Update("payment_status", payload.TransactionStatus)
	}
}
