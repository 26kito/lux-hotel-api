package service

import (
	"fmt"
	"log"
	"lux-hotel/entity"
	"lux-hotel/repository"
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

type UserService interface {
	Register(c echo.Context) error
	Login(c echo.Context) error
	GetBalance(c echo.Context) error
	TopUpBalance(c echo.Context) error
}

type userService struct {
	UserRepository repository.UserRepository
}

func NewUserService(userRepository repository.UserRepository) UserService {
	return &userService{UserRepository: userRepository}
}

func (us *userService) Register(c echo.Context) error {
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

	user, err := us.UserRepository.Register(request)

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

func (us *userService) Login(c echo.Context) error {
	var request entity.UserLoginPayload

	c.Bind(&request)

	// Validate the request payload
	if err := validateLoginPayload(request); err != nil {
		errCode, _ := strconv.Atoi(err.Error()[:3])
		errMessage := err.Error()[6:]

		return c.JSON(errCode, entity.ResponseError{
			Status:  errCode,
			Message: errMessage,
		})
	}

	user, err := us.UserRepository.Login(request)

	if err != nil {
		errCode, _ := strconv.Atoi(err.Error()[:3])
		errMessage := err.Error()[6:]

		return c.JSON(errCode, entity.ResponseError{
			Status:  errCode,
			Message: errMessage,
		})
	}

	// Generate JWT token
	tokenString, err := generateJWTToken(user)

	if err != nil {
		return c.JSON(500, entity.ResponseError{
			Status:  500,
			Message: "Failed to generate token",
		})
	}

	return c.JSON(200, entity.ResponseOK{
		Status:  200,
		Message: "User logged in successfully",
		Data: map[string]string{
			"token": tokenString,
		},
	})
}

func (us *userService) GetBalance(c echo.Context) error {
	userID := c.Get("user").(jwt.MapClaims)["user_id"].(float64)

	balance, err := us.UserRepository.GetBalance(int(userID))

	if err != nil {
		errCode, _ := strconv.Atoi(err.Error()[:3])
		errMessage := err.Error()[6:]

		return c.JSON(errCode, entity.ResponseError{
			Status:  errCode,
			Message: errMessage,
		})
	}

	return c.JSON(200, entity.ResponseOK{
		Status:  200,
		Message: "User balance retrieved successfully",
		Data: map[string]float64{
			"balance": balance,
		},
	})
}

func (us *userService) TopUpBalance(c echo.Context) error {
	userID := c.Get("user").(jwt.MapClaims)["user_id"].(float64)

	var request entity.UserTopUpBalancePayload

	c.Bind(&request)

	if err := validateTopUpPayload(request); err != nil {
		errCode, _ := strconv.Atoi(err.Error()[:3])
		errMessage := err.Error()[6:]

		return c.JSON(errCode, entity.ResponseError{
			Status:  errCode,
			Message: errMessage,
		})
	}

	response, err := us.UserRepository.TopUpBalance(int(userID), request)

	if err != nil {
		errCode, _ := strconv.Atoi(err.Error()[:3])
		errMessage := err.Error()[6:]

		return c.JSON(errCode, entity.ResponseError{
			Status:  errCode,
			Message: errMessage,
		})
	}

	return c.JSON(200, entity.ResponseOK{
		Status:  200,
		Message: "User balance topped up successfully",
		Data: map[string]interface{}{
			"order_id": response.OrderID,
		},
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

func validateLoginPayload(request entity.UserLoginPayload) error {
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
		return fmt.Errorf("400 | password is not valid")
	}

	return nil
}

func generateJWTToken(user *entity.User) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.UserID,
		"email":   user.Email,
		"exp":     time.Now().Add(time.Hour * 1).Unix(),
	})

	tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SECRET_KEY")))

	if err != nil {
		log.Println(err)
		return "", err
	}

	return tokenString, nil
}

func validateTopUpPayload(request entity.UserTopUpBalancePayload) error {
	if request.Amount == 0 {
		return fmt.Errorf("400 | amount is required")
	}

	if request.Amount <= 500000.00 {
		return fmt.Errorf("400 | top-up amount must be at least 500.000,00 IDR")
	}

	return nil
}
