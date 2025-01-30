package entity

import "time"

type User struct {
	UserID      uint      `gorm:"primaryKey"`
	FirstName   string    `gorm:"not null" json:"first_name"`
	LastName    string    `gorm:"type:varchar(100)" json:"last_name"`
	Email       string    `gorm:"unique" json:"email"`
	Password    string    `gorm:"type:varchar(255);not null" json:"-"`
	PhoneNumber string    `gorm:"type:varchar(15)" json:"phone_number"`
	Balance     float64   `gorm:"type:decimal(10,2);default:0" json:"balance"`
	CreatedAt   time.Time `gorm:"type:timestamp" json:"created_at"`
}

type UserRegisterPayload struct {
	FirstName   string `json:"first_name" form:"first_name" query:"first_name"`
	LastName    string `json:"last_name" form:"last_name" query:"last_name"`
	Email       string `json:"email" form:"email" query:"email"`
	Password    string `json:"password" form:"password" query:"password"`
	PhoneNumber string `json:"phone_number" form:"phone_number" query:"phone_number"`
}

type UserLoginPayload struct {
	Email    string `json:"email" form:"email" query:"email"`
	Password string `json:"password" form:"password" query:"password"`
}

type GetUserByEmailPayload struct {
	Email string `json:"email" form:"email" query:"email"`
}

type TopUpTransaction struct {
	ID                uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID            uint      `gorm:"not null" json:"user_id"`
	OrderID           string    `gorm:"unique;not null" json:"order_id"`
	Amount            float64   `gorm:"type:decimal(10,2);not null" json:"amount"`
	TransactionStatus string    `gorm:"type:varchar(20);default:'pending'" json:"transaction_status"`
	CreatedAt         time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt         time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

type UserTopUpBalancePayload struct {
	Amount float64 `json:"amount" form:"amount" query:"amount"`
}
