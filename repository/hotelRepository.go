package repository

import (
	"fmt"
	"lux-hotel/entity"
	"lux-hotel/utils"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type HotelRepository interface {
	GetHotelList() ([]entity.Hotel, error)
	GetHotelDetail(id int) (entity.Hotel, error)
	Booking(userID, hotelID int, request entity.BookingRequest) (*entity.Booking, error)
	Payment(payload entity.BookingPaymentPayload) (*entity.MidtransResponse, error)
}

type hotelRepository struct {
	DB *gorm.DB
}

func NewHotelRepository(db *gorm.DB) HotelRepository {
	return &hotelRepository{DB: db}
}

func (hr *hotelRepository) GetHotelList() ([]entity.Hotel, error) {
	var hotels []entity.Hotel

	result := hr.DB.Preload("Rooms").Find(&hotels)

	if result.Error != nil {
		return nil, fmt.Errorf("500 | %v", result.Error)
	}

	return hotels, nil
}

func (hr *hotelRepository) GetHotelDetail(id int) (entity.Hotel, error) {
	var hotel entity.Hotel

	result := hr.DB.Preload("Rooms").First(&hotel, id)

	if result.Error != nil {
		if result.Error.Error() == "record not found" {
			return hotel, fmt.Errorf("404 | Hotel not found")
		}

		return hotel, fmt.Errorf("500 | %v", result.Error)
	}

	return hotel, nil
}

func (hr *hotelRepository) Booking(userID, hotelID int, request entity.BookingRequest) (*entity.Booking, error) {
	var hotel entity.Hotel
	var room entity.Room
	var user entity.User

	// Parse the CheckIn and CheckOut strings into time.Time
	checkIn, err := time.Parse("2006-01-02", request.CheckIn)
	if err != nil {
		return nil, fmt.Errorf("invalid check-in date format")
	}

	checkOut, err := time.Parse("2006-01-02", request.CheckOut)
	if err != nil {
		return nil, fmt.Errorf("invalid check-out date format")
	}

	// Calculate the difference in days
	if checkOut.Before(checkIn) {
		return nil, fmt.Errorf("check-out date cannot be before check-in date")
	}

	// Total days is the difference in time divided by 24 hours
	totalDays := int(checkOut.Sub(checkIn).Hours() / 24)

	result := hr.DB.First(&hotel, hotelID)

	if result.Error != nil {
		if result.Error.Error() == "record not found" {
			return nil, fmt.Errorf("404 | Hotel not found")
		}

		return nil, fmt.Errorf("500 | %v", result.Error)
	}

	result = hr.DB.Where("hotel_id = ? AND id = ?", hotelID, request.RoomID).First(&room)

	if result.Error != nil {
		if result.Error.Error() == "record not found" {
			return nil, fmt.Errorf("404 | Room not found")
		}

		return nil, fmt.Errorf("500 | %v", result.Error)
	}

	result = hr.DB.First(&user, userID)

	if result.Error != nil {
		if result.Error.Error() == "record not found" {
			return nil, fmt.Errorf("404 | User not found")
		}

		return nil, fmt.Errorf("500 | %v", result.Error)
	}

	orderID := fmt.Sprintf("BKNG-%d%s", userID, uuid.New().String())
	bookingCode := fmt.Sprintf("%s%d%d", time.Now().Format("20060102"), hotelID, request.RoomID)

	booking := entity.Booking{
		OrderID:       orderID,
		BookingCode:   bookingCode,
		GuestID:       user.UserID,
		HotelID:       hotel.ID,
		RoomID:        request.RoomID,
		CheckIn:       checkIn.Format("2006-01-02"),
		CheckOut:      checkOut.Format("2006-01-02"),
		TotalDays:     totalDays,
		TotalPrice:    float64(totalDays) * room.Price,
		BookingStatus: "pending",
	}

	result = hr.DB.Create(&booking)

	if result.Error != nil {
		return nil, fmt.Errorf("500 | %v", result.Error)
	}

	return &booking, nil
}

func (hr *hotelRepository) Payment(payload entity.BookingPaymentPayload) (*entity.MidtransResponse, error) {
	var user entity.User
	var booking entity.Booking
	var payment entity.Payment

	result := hr.DB.Where("order_id = ?", payload.OrderID).First(&booking)

	if result.Error != nil {
		if result.Error.Error() == "record not found" {
			return nil, fmt.Errorf("404 | Booking not found")
		}

		return nil, fmt.Errorf("500 | %v", result.Error)
	}

	if booking.BookingStatus != "pending" {
		return nil, fmt.Errorf("400 | Booking has been paid")
	}

	transaction, err := utils.MidtransTransactionStatusHandler(payload.OrderID)

	if err != nil {
		return nil, fmt.Errorf("500 | %v", err)
	}

	if transaction.TransactionStatus == "pending" {
		return transaction, nil
	}

	result = hr.DB.Where("user_id = ?", booking.GuestID).First(&user)

	if result.Error != nil {
		if result.Error.Error() == "record not found" {
			return nil, fmt.Errorf("404 | User not found")
		}

		return nil, fmt.Errorf("500 | %v", result.Error)
	}

	// Prepare the struct for Midtrans
	midtransPayload := entity.MidtransPaymentPayload{
		PaymentType: "bank_transfer",
		TransactionDetail: struct {
			OrderID     string  `json:"order_id"`
			GrossAmount float64 `json:"gross_amount"`
		}{
			OrderID:     payload.OrderID,
			GrossAmount: booking.TotalPrice,
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
				Price:    booking.TotalPrice,
				Quantity: 1,
				Name:     "Hotel Booking",
			},
		},
		BankTransfer: struct {
			Bank string `json:"bank"`
		}{
			Bank: payload.PaymentMethod,
		},
	}

	response, err := utils.MidtransPaymentHandler(midtransPayload)

	if err != nil {
		return nil, fmt.Errorf("500 | %v", err)
	}

	payment = entity.Payment{
		PaymentID:       response.TransactionID,
		OrderID:         payload.OrderID,
		UserID:          booking.GuestID,
		TotalAmount:     booking.TotalPrice,
		TransactionType: "booking",
		PaymentDate:     nil,
		PaymentStatus:   "pending",
		PaymentMethod:   response.PaymentType + " - " + response.VANumbers[0].Bank,
	}

	result = hr.DB.Create(&payment)

	if result.Error != nil {
		return nil, fmt.Errorf("500 | %v", result.Error)
	}

	return response, nil
}
