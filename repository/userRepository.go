package repository

import (
	"fmt"
	"log"
	"lux-hotel/entity"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserRepository interface {
	Register(entity.UserRegisterPayload) (*entity.User, error)
	Login(entity.UserLoginPayload) (*entity.User, error)
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
