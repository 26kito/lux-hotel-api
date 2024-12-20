package repository

import (
	"fmt"
	"log"
	"lux-hotel/entity"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserRepository interface {
	Register(entity.UserRegisterPayload) (*entity.User, error)
	Login(entity.UserLoginPayload) (*entity.User, error)
	GetBalance(int) (float64, error)
	TopUpBalance(int, entity.UserTopUpBalancePayload) (*entity.TopUpTransaction, error)
	GetBookHistory(int) ([]entity.BookingHistoryResponse, error)
}

type userRepository struct {
	DB *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{DB: db}
}

func (ur *userRepository) Register(request entity.UserRegisterPayload) (*entity.User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(request.Password), bcrypt.DefaultCost)

	request.Password = string(hashedPassword)

	if err != nil {
		log.Println(err)
		return nil, fmt.Errorf("500 | internal server error")
	}

	if emailExists := ur.DB.Where("email = ?", request.Email).First(&entity.User{}); emailExists.RowsAffected > 0 {
		return nil, fmt.Errorf("409 | email already exists")
	}

	if phoneNumberExists := ur.DB.Where("phone_number = ?", request.PhoneNumber).First(&entity.User{}); phoneNumberExists.RowsAffected > 0 {
		return nil, fmt.Errorf("409 | phone number already exists")
	}

	user := entity.User{
		FirstName:   request.FirstName,
		LastName:    request.LastName,
		Email:       request.Email,
		Password:    request.Password,
		PhoneNumber: request.PhoneNumber,
	}

	result := ur.DB.Create(&user)

	if result.Error != nil {
		log.Println(result.Error)
		return nil, fmt.Errorf("500 | internal server error")
	}

	return &user, nil
}

func (ur *userRepository) Login(request entity.UserLoginPayload) (*entity.User, error) {
	var user entity.User

	result := ur.DB.Where("email = ?", request.Email).First(&user)

	if result.Error != nil {
		log.Println(result.Error)
		return nil, fmt.Errorf("404 | user not found")
	}

	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(request.Password))

	if err != nil {
		log.Println(err)
		return nil, fmt.Errorf("401 | unauthorized")
	}

	return &user, nil
}

func (ur *userRepository) GetBalance(userID int) (float64, error) {
	var user entity.User

	result := ur.DB.Where("user_id = ?", userID).First(&user)

	if result.Error != nil {
		log.Println(result.Error)
		return 0, fmt.Errorf("404 | user not found")
	}

	return user.Balance, nil
}

func (ur *userRepository) TopUpBalance(userID int, request entity.UserTopUpBalancePayload) (*entity.TopUpTransaction, error) {
	orderID := fmt.Sprintf("TPUP-%s", uuid.New().String())

	topup := ur.createTopupEntity(uint(userID), orderID, request.Amount)

	insertTopup := ur.DB.Save(&topup)

	if insertTopup.Error != nil {
		log.Println(insertTopup.Error)
		return nil, fmt.Errorf("500 | internal server error")
	}

	return &topup, nil
}

func (ur *userRepository) GetBookHistory(userID int) ([]entity.BookingHistoryResponse, error) {
	var historyBook []entity.BookingHistoryResponse

	result := ur.DB.Table("bookings").
		Select("bookings.order_id, bookings.booking_code, hotels.name AS hotel_name, rooms.room_number, bookings.check_in, bookings.check_out, bookings.total_days, bookings.total_price, bookings.created_at AS booking_date, bookings.booking_status").
		Joins("JOIN hotels ON bookings.hotel_id = hotels.id").
		Joins("JOIN rooms ON bookings.room_id = rooms.id").
		Where("bookings.guest_id = ?", userID).
		Scan(&historyBook)

	if result.Error != nil {
		log.Println(result.Error)
		return nil, fmt.Errorf("500 | internal server error")
	}

	return historyBook, nil
}

func (ur *userRepository) createTopupEntity(userID uint, orderID string, amount float64) entity.TopUpTransaction {
	return entity.TopUpTransaction{
		UserID:  userID,
		OrderID: orderID,
		Amount:  amount,
	}
}
