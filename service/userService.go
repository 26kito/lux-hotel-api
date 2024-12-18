package service

import (
	"fmt"
	"lux-hotel/entity"
	"lux-hotel/repository"
	"strconv"

	"github.com/labstack/echo/v4"
)

type UserService interface {
	Register(c echo.Context) error
}

type userService struct {
	UserRepository repository.UserRepository
}

func NewUserService(userRepository repository.UserRepository) UserService {
	return &userService{UserRepository: userRepository}
}

func (s *userService) Register(c echo.Context) error {
	var request entity.UserRegisterPayload

	c.Bind(&request)

	if err := validateRegisterPayload(request); err != nil {
		errCode, _ := strconv.Atoi(err.Error()[:3])
		errMessage := err.Error()[6:]

		return c.JSON(errCode, entity.ResponseError{
			Status:  errCode,
			Message: errMessage,
		})
	}

	user, err := s.UserRepository.Register(request)

	if err != nil {
		errCode, _ := strconv.Atoi(err.Error()[:3])
		errMessage := err.Error()[6:]

		return c.JSON(errCode, entity.ResponseError{
			Status:  errCode,
			Message: errMessage,
		})
	}

	return c.JSON(201, entity.ResponseOK{
		Status:  201,
		Message: "User registered successfully",
		Data:    user,
	})
}

func validateRegisterPayload(request entity.UserRegisterPayload) error {
	if request.FirstName == "" {
		return fmt.Errorf("400 | first name is required")
	}

	if len(request.FirstName) < 3 {
		return fmt.Errorf("400 | first name must be at least 3 characters")
	}

	if request.Email == "" {
		return fmt.Errorf("400 | email is required")
	}

	if len(request.Email) < 8 {
		return fmt.Errorf("400 | email is not valid")
	}

	if request.Password == "" {
		return fmt.Errorf("400 | password is required")
	}

	if len(request.Password) < 8 {
		return fmt.Errorf("400 | password must be at least 8 characters")
	}

	if request.PhoneNumber == "" {
		return fmt.Errorf("400 | phone number is required")
	}

	if request.PhoneNumber[:2] != "08" && request.PhoneNumber[:3] != "628" {
		return fmt.Errorf("400 | phone number is not valid")
	}

	if len(request.PhoneNumber) < 8 {
		return fmt.Errorf("400 | phone number is not valid")
	}

	return nil
}
