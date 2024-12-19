package repository

import (
	"fmt"
	"lux-hotel/entity"
	"lux-hotel/utils"
	"strings"

	"gorm.io/gorm"
)

type PaymentRepository interface {
	Payment(entity.PaymentPayload) (*entity.MidtransResponse, error)
}

type paymentRepository struct {
	DB *gorm.DB
}

func NewPaymentRepository(db *gorm.DB) PaymentRepository {
	return &paymentRepository{DB: db}
}

func (pr *paymentRepository) Payment(payload entity.PaymentPayload) (*entity.MidtransResponse, error) {
	var response *entity.MidtransResponse
	var err error

	// Determine transaction type by OrderID prefix
	if strings.HasPrefix(payload.OrderID, "TPUP") {
		response, err = pr.handleTopUpPayment(payload)
	} else if strings.HasPrefix(payload.OrderID, "BKNG") {
		response, err = pr.handleBookingPayment(payload)
	} else {
		return nil, fmt.Errorf("400 | Invalid OrderID format")
	}

	if err != nil {
		return nil, err
	}

	return response, nil
}

func (pr *paymentRepository) handleTopUpPayment(payload entity.PaymentPayload) (*entity.MidtransResponse, error) {
	topup, topupErr := pr.getTopupTransactionByOrderID(payload.OrderID)
	if topupErr != nil {
		return nil, topupErr
	}

	transaction, err := utils.MidtransTransactionStatusHandler(payload.OrderID)
	if err != nil {
		return nil, fmt.Errorf("500 | %v", err)
	}

	if transaction.TransactionStatus == "pending" {
		return transaction, nil
	}

	user, userErr := pr.getUserByID(topup.UserID)

	if userErr != nil {
		return nil, userErr
	}

	midtransPayload := pr.prepareMidtransPayload(payload, user, topup.Amount, "topup balance")

	response, responseErr := utils.MidtransPaymentHandler(midtransPayload)

	if responseErr != nil {
		return nil, fmt.Errorf("500 | %v", responseErr)
	}

	payment := pr.createPaymentEntity(response, payload.OrderID, user.UserID, topup.Amount, "topup balance")

	if err := pr.savePayment(payment); err != nil {
		return nil, err
	}

	return response, nil
}

func (pr *paymentRepository) handleBookingPayment(payload entity.PaymentPayload) (*entity.MidtransResponse, error) {
	booking, bookingErr := pr.getBookingByOrderID(payload.OrderID)
	if bookingErr != nil {
		return nil, bookingErr
	}

	if err := pr.validateBookingStatus(booking); err != nil {
		return nil, err
	}

	transaction, err := utils.MidtransTransactionStatusHandler(payload.OrderID)
	if err != nil {
		return nil, fmt.Errorf("500 | %v", err)
	}

	if transaction.TransactionStatus == "pending" {
		return transaction, nil
	}

	user, err := pr.getUserByID(booking.GuestID)
	if err != nil {
		return nil, err
	}

	midtransPayload := pr.prepareMidtransPayload(payload, user, booking.TotalPrice, "hotel booking")
	response, err := utils.MidtransPaymentHandler(midtransPayload)

	if err != nil {
		return nil, fmt.Errorf("500 | %v", err)
	}

	payment := pr.createPaymentEntity(response, payload.OrderID, user.UserID, booking.TotalPrice, "hotel booking")

	if err := pr.savePayment(payment); err != nil {
		return nil, err
	}

	return response, nil
}

func (pr *paymentRepository) getBookingByOrderID(orderID string) (*entity.Booking, error) {
	var booking entity.Booking

	result := pr.DB.Where("order_id = ?", orderID).First(&booking)

	if result.Error != nil {
		if result.Error.Error() == "record not found" {
			return nil, fmt.Errorf("404 | Booking not found")
		}

		return nil, fmt.Errorf("500 | %v", result.Error)
	}

	return &booking, nil
}

func (pr *paymentRepository) validateBookingStatus(booking *entity.Booking) error {
	if booking.BookingStatus != "pending" {
		return fmt.Errorf("400 | Booking has been paid")
	}

	return nil
}

func (pr *paymentRepository) getTopupTransactionByOrderID(orderID string) (*entity.TopUpTransaction, error) {
	var topup entity.TopUpTransaction

	result := pr.DB.Where("order_id = ?", orderID).First(&topup)

	if result.Error != nil {
		if result.Error.Error() == "record not found" {
			return nil, fmt.Errorf("404 | Booking not found")
		}

		return nil, fmt.Errorf("500 | %v", result.Error)
	}

	return &topup, nil
}

func (pr *paymentRepository) getUserByID(userID uint) (*entity.User, error) {
	var user entity.User

	result := pr.DB.Where("user_id = ?", userID).First(&user)

	if result.Error != nil {
		if result.Error.Error() == "record not found" {
			return nil, fmt.Errorf("404 | User not found")
		}

		return nil, fmt.Errorf("500 | %v", result.Error)
	}

	return &user, nil
}

func (pr *paymentRepository) prepareMidtransPayload(payload entity.PaymentPayload, user *entity.User, amount float64, txName string) entity.MidtransPaymentPayload {
	return entity.MidtransPaymentPayload{
		PaymentType: "bank_transfer",
		TransactionDetail: struct {
			OrderID     string  `json:"order_id"`
			GrossAmount float64 `json:"gross_amount"`
		}{
			OrderID:     payload.OrderID,
			GrossAmount: amount,
		},
		CustomerDetail: struct {
			Email     string `json:"email"`
			FirstName string `json:"first_name"`
			LastName  string `json:"last_name"`
			Phone     string `json:"phone"`
		}{
			Email:     user.Email,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			Phone:     user.PhoneNumber,
		},
		ItemDetails: []struct {
			ID       string  `json:"id"`
			Price    float64 `json:"price"`
			Quantity int     `json:"quantity"`
			Name     string  `json:"name"`
		}{
			{
				ID:       payload.OrderID,
				Price:    amount,
				Quantity: 1,
				Name:     txName,
			},
		},
		BankTransfer: struct {
			Bank string `json:"bank"`
		}{
			Bank: payload.PaymentMethod,
		},
	}
}

func (pr *paymentRepository) createPaymentEntity(response *entity.MidtransResponse, orderID string, userID uint, amount float64, transactionType string) entity.Payment {
	return entity.Payment{
		PaymentID:       response.TransactionID,
		OrderID:         orderID,
		UserID:          userID,
		TotalAmount:     amount,
		TransactionType: transactionType,
		PaymentDate:     nil,
		PaymentStatus:   "pending",
		PaymentMethod:   response.PaymentType + " - " + response.VANumbers[0].Bank,
	}
}

func (pr *paymentRepository) savePayment(payment entity.Payment) error {
	result := pr.DB.Create(&payment)

	if result.Error != nil {
		return fmt.Errorf("500 | %v", result.Error)
	}

	return nil
}
