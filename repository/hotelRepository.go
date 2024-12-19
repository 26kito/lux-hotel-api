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
	// Parse checkin and checkout
	checkIn, checkOut, parseDateError := hr.parseBookingDates(request.CheckIn, request.CheckOut)

	if parseDateError != nil {
		return nil, parseDateError
	}

	if checkOut.Before(checkIn) {
		return nil, fmt.Errorf("check-out date cannot be before check-in date")
	}

	// Total days is the difference in time divided by 24 hours
	totalDays := int(checkOut.Sub(checkIn).Hours() / 24)

	// Get hotel
	hotel, err := hr.getHotelByID(hotelID)
	if err != nil {
		return nil, err
	}

	// Get room
	room, err := hr.getHotelRoom(uint(hotelID), request.RoomID)
	if err != nil {
		return nil, err
	}

	// Get user
	user, err := hr.getUserByID(uint(userID))

	if err != nil {
		return nil, err
	}

	orderID := fmt.Sprintf("BKNG-%d%s", userID, uuid.New().String())
	bookingCode := fmt.Sprintf("%s%d%d", time.Now().Format("20060102"), hotelID, request.RoomID)
	room.Price = utils.AdjustPrice(room.Price, checkIn)
	totalPrice := float64(totalDays) * room.Price

	booking := hr.createBookingEntity(orderID, bookingCode, *user, *hotel, *room, checkIn, checkOut, totalDays, totalPrice)

	insertBooking := hr.DB.Save(&booking)

	if insertBooking.Error != nil {
		return nil, fmt.Errorf("500 | %v", insertBooking.Error)
	}

	return &booking, nil
}

func (hr *hotelRepository) parseBookingDates(checkInStr, checkOutStr string) (time.Time, time.Time, error) {
	checkIn, err := time.Parse("2006-01-02", checkInStr)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("500 | invalid check-in date format")
	}

	checkOut, err := time.Parse("2006-01-02", checkOutStr)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("500 | invalid check-out date format")
	}

	return checkIn, checkOut, nil
}

func (hr *hotelRepository) getUserByID(userID uint) (*entity.User, error) {
	var user entity.User

	result := hr.DB.Where("user_id = ?", userID).First(&user)

	if result.Error != nil {
		if result.Error.Error() == "record not found" {
			return nil, fmt.Errorf("404 | User not found")
		}

		return nil, fmt.Errorf("500 | %v", result.Error)
	}

	return &user, nil
}

func (hr *hotelRepository) getHotelByID(hotelID int) (*entity.Hotel, error) {
	var hotel entity.Hotel

	result := hr.DB.First(&hotel, hotelID)

	if result.Error != nil {
		if result.Error.Error() == "record not found" {
			return nil, fmt.Errorf("404 | Hotel not found")
		}

		return nil, fmt.Errorf("500 | %v", result.Error)
	}

	return &hotel, nil
}

func (hr *hotelRepository) getHotelRoom(hotelID, roomID uint) (*entity.Room, error) {
	var room entity.Room

	result := hr.DB.Where("hotel_id = ? AND id = ?", hotelID, roomID).First(&room)

	if result.Error != nil {
		if result.Error.Error() == "record not found" {
			return nil, fmt.Errorf("404 | Room not found")
		}

		return nil, fmt.Errorf("500 | %v", result.Error)
	}

	if room.Status == "occupied" {
		return nil, fmt.Errorf("400 | Room is occupied")
	}

	return &room, nil
}

func (hr *hotelRepository) createBookingEntity(orderID string, bookingCode string, user entity.User, hotel entity.Hotel, room entity.Room, checkIn time.Time, checkOut time.Time, totalDays int, totalPrice float64) entity.Booking {
	return entity.Booking{
		OrderID:       orderID,
		BookingCode:   bookingCode,
		GuestID:       user.UserID,
		HotelID:       hotel.ID,
		RoomID:        room.ID,
		CheckIn:       checkIn.Format("2006-01-02"),
		CheckOut:      checkOut.Format("2006-01-02"),
		TotalDays:     totalDays,
		TotalPrice:    totalPrice,
		BookingStatus: "pending",
	}
}
