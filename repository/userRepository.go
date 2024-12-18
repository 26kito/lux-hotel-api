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
}

type userRepository struct {
	DB *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{DB: db}
}

func (r *userRepository) Register(request entity.UserRegisterPayload) (*entity.User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(request.Password), bcrypt.DefaultCost)

	request.Password = string(hashedPassword)

	if err != nil {
		log.Println(err)
		return nil, fmt.Errorf("500 | internal server error")
	}

	user := entity.User{
		FirstName:   request.FirstName,
		LastName:    request.LastName,
		Email:       request.Email,
		Password:    request.Password,
		PhoneNumber: request.PhoneNumber,
	}

	result := r.DB.Create(&user)

	if result.Error != nil {
		log.Println(result.Error)
		return nil, fmt.Errorf("500 | internal server error")
	}

	return &user, nil
}
