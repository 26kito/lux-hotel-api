package repository

import (
	"fmt"
	"lux-hotel/entity"
	"lux-hotel/utils"
	"strings"
	"time"

	"gorm.io/gorm"
)

type PaymentRepository interface {
	Payment(int, entity.PaymentPayload) (*entity.PaymentResponse, error)
}

type paymentRepository struct {
	DB *gorm.DB
}

func NewPaymentRepository(db *gorm.DB) PaymentRepository {
	return &paymentRepository{DB: db}
}

func (pr *paymentRepository) Payment(userID int, payload entity.PaymentPayload) (*entity.PaymentResponse, error) {
	var response *entity.PaymentResponse
	var err error

	// Determine transaction type by OrderID prefix
	if strings.HasPrefix(payload.OrderID, "TPUP") {
		response, err = pr.handleTopUpPayment(userID, payload)
	} else if strings.HasPrefix(payload.OrderID, "BKNG") {
		response, err = pr.handleBookingPayment(userID, payload)
	} else {
		return nil, fmt.Errorf("400 | Invalid OrderID format")
	}

	if err != nil {
		return nil, err
	}

	return response, nil
}

func (pr *paymentRepository) handleTopUpPayment(userID int, payload entity.PaymentPayload) (*entity.PaymentResponse, error) {
	// Fetch the top-up transaction details by order ID
	topup, topupErr := pr.getTopupTransactionByOrderID(payload.OrderID)
	if topupErr != nil {
		return nil, topupErr
	}

	if topup.UserID != uint(userID) {
		return nil, fmt.Errorf("401 | Unauthorized access")
	}

	if topup.TransactionStatus == "cancel" || topup.TransactionStatus == "failed" {
		return nil, fmt.Errorf("400 | Top-up transaction has been %s", topup.TransactionStatus)
	}

	// Check the transaction status from Midtrans
	transaction, transactionErr := utils.MidtransTransactionStatusHandler(payload.OrderID)
	if transactionErr != nil {
		return nil, fmt.Errorf("500 | %v", transactionErr)
	}

	// Return pending transaction details
	if transaction.TransactionStatus == "pending" {
		return &entity.PaymentResponse{
			TransactionID:     transaction.TransactionID,
			TransactionStatus: transaction.TransactionStatus,
			Amount:            utils.StringToFloat64(transaction.GrossAmount),
			PaymentType:       transaction.PaymentType,
			Bank:              transaction.VANumbers[0].Bank,
			VANumber:          transaction.VANumbers[0].VANumber,
		}, nil
	}

	// Fetch the user associated with the top-up
	user, userErr := pr.getUserByID(topup.UserID)
	if userErr != nil {
		return nil, userErr
	}

	// Prepare the Midtrans payload for the top-up
	midtransPayload := pr.prepareMidtransPayload(payload, user, topup.Amount, "topup balance")

	// Handle the payment via Midtrans
	response, responseErr := utils.MidtransPaymentHandler(midtransPayload)
	if responseErr != nil {
		return nil, fmt.Errorf("500 | %v", responseErr)
	}

	transactionID := fmt.Sprintf("TRX-%d", time.Now().Unix())
	paymentMethod := response.PaymentType + " - " + response.VANumbers[0].Bank

	// Create and save the payment entity
	payment := pr.createPaymentEntity(transactionID, payload.OrderID, user.UserID, topup.Amount, "topup balance", nil, "pending", paymentMethod)
	if err := pr.savePayment(payment); err != nil {
		return nil, err
	}

	// Return PaymentResponse
	return &entity.PaymentResponse{
		TransactionID:     payment.PaymentID,
		TransactionStatus: payment.PaymentStatus,
		Amount:            payment.TotalAmount,
		PaymentType:       payment.TransactionType,
		Bank:              response.VANumbers[0].Bank,
		VANumber:          response.VANumbers[0].VANumber,
	}, nil
}

func (pr *paymentRepository) handleBookingPayment(userID int, payload entity.PaymentPayload) (*entity.PaymentResponse, error) {
	// Fetch booking details by order ID
	booking, bookingErr := pr.getBookingByOrderID(payload.OrderID)
	if bookingErr != nil {
		return nil, bookingErr
	}

	if booking.GuestID != uint(userID) {
		return nil, fmt.Errorf("401 | Unauthorized access")
	}

	// Validate booking status
	if err := pr.validateBookingStatus(booking); err != nil {
		return nil, err
	}

	// Retrieve user details
	user, err := pr.getUserByID(booking.GuestID)
	if err != nil {
		return nil, err
	}

	switch payload.PaymentMethod {
	case "wallet":
		return pr.handleWalletPayment(payload, booking, user)
	default:
		return pr.handleBankPayment(payload, booking, user)
	}
}

func (pr *paymentRepository) handleWalletPayment(payload entity.PaymentPayload, booking *entity.Booking, user *entity.User) (*entity.PaymentResponse, error) {
	// Ensure user balance is sufficient
	if user.Balance < booking.TotalPrice {
		return nil, fmt.Errorf("400 | Insufficient balance")
	}

	// Deduct balance and save
	user.Balance -= booking.TotalPrice
	if err := pr.DB.Save(&user).Error; err != nil {
		return nil, fmt.Errorf("500 | Failed to update user balance: %v", err)
	}

	// Update booking status to "settlement"
	booking.BookingStatus = "settlement"
	if err := pr.DB.Save(&booking).Error; err != nil {
		return nil, fmt.Errorf("500 | Failed to update booking status: %v", err)
	}

	transactionID := fmt.Sprintf("TRX-%d", time.Now().Unix())
	paymentDate, _ := time.Parse("2006-01-02", time.Now().Format("2006-01-02"))
	// Create and save payment entity
	payment := pr.createPaymentEntity(transactionID, payload.OrderID, user.UserID, booking.TotalPrice, "hotel booking", &paymentDate, "settlement", payload.PaymentMethod)
	payment.PaymentStatus = "settlement"
	payment.PaymentMethod = "wallet"

	if err := pr.savePayment(payment); err != nil {
		return nil, fmt.Errorf("500 | Failed to save payment record: %v", err)
	}

	// Return PaymentResponse
	return &entity.PaymentResponse{
		TransactionID:     payment.PaymentID,
		TransactionStatus: payment.PaymentStatus,
		Amount:            payment.TotalAmount,
		PaymentType:       payment.TransactionType,
	}, nil
}

func (pr *paymentRepository) handleBankPayment(payload entity.PaymentPayload, booking *entity.Booking, user *entity.User) (*entity.PaymentResponse, error) {
	midtransPayload := pr.prepareMidtransPayload(payload, user, booking.TotalPrice, "hotel booking")
	response, err := utils.MidtransPaymentHandler(midtransPayload)
	if err != nil {
		return nil, fmt.Errorf("500 | %v", err)
	}

	paymentMethod := response.PaymentType + " - " + response.VANumbers[0].Bank

	transactionID := fmt.Sprintf("TRX-%d", time.Now().Unix())
	// Create and save payment entity
	payment := pr.createPaymentEntity(transactionID, payload.OrderID, user.UserID, booking.TotalPrice, "hotel booking", nil, "pending", paymentMethod)
	if err := pr.savePayment(payment); err != nil {
		return nil, err
	}

	// Return PaymentResponse
	return &entity.PaymentResponse{
		TransactionID:     payment.PaymentID,
		TransactionStatus: payment.PaymentStatus,
		Amount:            payment.TotalAmount,
		PaymentType:       payment.TransactionType,
		Bank:              response.VANumbers[0].Bank,
		VANumber:          response.VANumbers[0].VANumber,
	}, nil
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
			OrderID     string `json:"order_id"`
			GrossAmount string `json:"gross_amount"`
		}{
			OrderID:     payload.OrderID,
			GrossAmount: fmt.Sprintf("%.2f", amount),
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
			ID       string `json:"id"`
			Price    string `json:"price"`
			Quantity int    `json:"quantity"`
			Name     string `json:"name"`
		}{
			{
				ID:       payload.OrderID,
				Price:    fmt.Sprintf("%.2f", amount),
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

func (pr *paymentRepository) createPaymentEntity(transactionID string, orderID string, userID uint, amount float64, transactionType string, paymentDate *time.Time, paymentStatus string, paymentMethod string) entity.Payment {
	return entity.Payment{
		PaymentID:       transactionID,
		OrderID:         orderID,
		UserID:          userID,
		TotalAmount:     amount,
		TransactionType: transactionType,
		PaymentDate:     paymentDate,
		PaymentStatus:   paymentStatus,
		PaymentMethod:   paymentMethod,
	}
}

func (pr *paymentRepository) savePayment(payment entity.Payment) error {
	result := pr.DB.Create(&payment)

	if result.Error != nil {
		return fmt.Errorf("500 | %v", result.Error)
	}

	return nil
}
