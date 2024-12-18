package entity

import "time"

type User struct {
	UserID      uint      `gorm:"primaryKey"`
	FirstName   string    `gorm:"not null" json:"first_name"`
	LastName    string    `gorm:"type:varchar(100)" json:"last_name"`
	Email       string    `gorm:"unique" json:"email"`
	Password    string    `gorm:"type:varchar(255);not null" json:"-"`
	PhoneNumber string    `gorm:"type:varchar(20)" json:"phone_number"`
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
