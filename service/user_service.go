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
	GetBookHistory(c echo.Context) error
}

type userService struct {
	UserRepository repository.UserRepository
}

func NewUserService(userRepository repository.UserRepository) UserService {
	return &userService{UserRepository: userRepository}
}

// Register handles user registration.
// @Summary Register a new user
// @Description Registers a new user in the system. It validates the input, checks for errors, and stores the user data in the database.
// @Tags user
// @Accept json
// @Produce json
// @Param user body entity.UserRegisterPayload true "User registration data"
// @Success 201 {object} entity.ResponseOK "User successfully registered"
// @Failure 400 {object} entity.ResponseError "Invalid registration data"
// @Failure 409 {object} entity.ResponseError "Email already exists"
// @Failure 500 {object} entity.ResponseError "Internal server error"
// @Router /api/users/register [post]
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

// Login handles user login and returns a JWT token.
// @Summary Login a user and return a JWT token
// @Description Logs the user in by validating their credentials and returning a JWT token for authentication.
// @Tags user
// @Accept json
// @Produce json
// @Param user body entity.UserLoginPayload true "User login data"
// @Success 200 {object} entity.ResponseOK "User logged in successfully"
// @Failure 400 {object} entity.ResponseError "Invalid login credentials"
// @Failure 401 {object} entity.ResponseError "Unauthorized access"
// @Failure 500 {object} entity.ResponseError "Internal server error"
// @Router /api/users/login [post]
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

// GetBalance retrieves the current balance of the logged-in user.
// @Summary Get the balance of the logged-in user
// @Description Retrieves the current balance of the user from the database based on the user ID obtained from the JWT token.
// @Tags user
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} entity.ResponseOK "User balance retrieved successfully"
// @Failure 400 {object} entity.ResponseError "Bad request"
// @Failure 401 {object} entity.ResponseError "Unauthorized access"
// @Failure 500 {object} entity.ResponseError "Internal server error"
// @Router /api/users/balance [get]
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

// TopUpBalance allows the logged-in user to top up their balance.
// @Summary Top up the balance of the logged-in user
// @Description Allows the user to top up their balance by providing the amount and other relevant information. The request must include a valid JWT token for authentication.
// @Tags user
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param user body entity.UserTopUpBalancePayload true "User top-up balance data"
// @Success 200 {object} entity.ResponseOK "User balance topped up successfully"
// @Failure 400 {object} entity.ResponseError "Invalid top-up data"
// @Failure 401 {object} entity.ResponseError "Unauthorized access"
// @Failure 500 {object} entity.ResponseError "Internal server error"
// @Router /api/users/balance/top-up [post]
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

// GetBookHistory retrieves the booking history of the logged-in user.
// @Summary Get user booking history
// @Description Fetches the booking history for the logged-in user based on the user ID extracted from the JWT token.
// @Tags user
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} entity.ResponseOK "User booking history retrieved successfully"
// @Failure 400 {object} entity.ResponseError "Bad request"
// @Failure 401 {object} entity.ResponseError "Unauthorized access"
// @Failure 500 {object} entity.ResponseError "Internal server error"
// @Router /api/users/book/history [get]
func (us *userService) GetBookHistory(c echo.Context) error {
	userID := c.Get("user").(jwt.MapClaims)["user_id"].(float64)

	history, err := us.UserRepository.GetBookHistory(int(userID))

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
		Message: "User history book retrieved successfully",
		Data:    history,
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
