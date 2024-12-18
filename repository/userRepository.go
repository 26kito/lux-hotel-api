package repository

import (
	"fmt"
	"log"
	"lux-hotel/entity"
	"lux-hotel/utils"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserRepository interface {
	Register(entity.UserRegisterPayload) (*entity.User, error)
	Login(entity.UserLoginPayload) (*entity.User, error)
	GetBalance(int) (float64, error)
	TopUpBalance(int, entity.UserTopUpBalancePayload) (*entity.MidtransResponse, error)
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

func (ur *userRepository) TopUpBalance(userID int, request entity.UserTopUpBalancePayload) (*entity.MidtransResponse, error) {
	var user entity.User
	var topUpTransaction entity.TopUpTransaction
	var payment entity.Payment

	result := ur.DB.Where("user_id = ?", userID).First(&user)

	if result.Error != nil {
		log.Println(result.Error)
		return nil, fmt.Errorf("404 | user not found")
	}

	orderID := fmt.Sprintf("TPUP-%s", uuid.New().String())

	topUpTransaction = entity.TopUpTransaction{
		UserID:  user.UserID,
		OrderID: orderID,
		Amount:  request.Amount,
	}

	result = ur.DB.Save(&topUpTransaction)

	if result.Error != nil {
		log.Println(result.Error)
		return nil, fmt.Errorf("500 | internal server error")
	}

	// Prepare the struct for Midtrans
	payload := entity.MidtransPaymentPayload{
		PaymentType: "bank_transfer",
		TransactionDetail: struct {
			OrderID     string  `json:"order_id"`
			GrossAmount float64 `json:"gross_amount"`
		}{
			OrderID:     orderID,
			GrossAmount: request.Amount,
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
				ID:       orderID,
				Price:    request.Amount,
				Quantity: 1,
				Name:     "Top-up Balance",
			},
		},
		BankTransfer: struct {
			Bank string `json:"bank"`
		}{
			Bank: request.BankTransfer,
		},
	}

	response, err := utils.MidtransPaymentHandler(payload)

	if err != nil {
		log.Println(err)
		return nil, fmt.Errorf("500 | internal server error")
	}

	payment = entity.Payment{
		PaymentID:       response.TransactionID,
		OrderID:         orderID,
		UserID:          user.UserID,
		TotalAmount:     request.Amount,
		TransactionType: "topup",
		PaymentDate:     nil,
		PaymentStatus:   response.TransactionStatus,
		PaymentMethod:   response.PaymentType + " - " + response.VANumbers[0].Bank,
	}

	result = ur.DB.Save(&payment)

	if result.Error != nil {
		log.Println(result.Error)
		return nil, fmt.Errorf("500 | internal server error")
	}

	return response, nil
}

// func midtransPaymentHandler(payload entity.MidtransPaymentPayload) (*entity.MidtransResponse, error) {
// 	var response entity.MidtransResponse

// 	client := resty.New()

// 	midtransServerKey := os.Getenv("MIDTRANS_SERVER_KEY")
// 	encodedKey := base64.StdEncoding.EncodeToString([]byte(midtransServerKey))

// 	url := os.Getenv("MIDTRANS_BASE_URL") + "/charge"

// 	resp, err := client.R().
// 		SetHeader("Accept", "application/json").
// 		SetHeader("Content-Type", "application/json").
// 		SetHeader("Authorization", fmt.Sprintf("Basic %s", encodedKey)).
// 		SetBody(payload).
// 		Post(url)

// 	if err != nil {
// 		log.Println(err)
// 		return nil, fmt.Errorf("500 | %v", err)
// 	}

// 	if err := json.Unmarshal(resp.Body(), &response); err != nil {
// 		log.Println("Error unmarshalling response:", err)
// 		return nil, err
// 	}

// 	return &response, nil
// }
