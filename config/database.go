package config

import (
	"fmt"
	"log"
	"os"

	"lux-hotel/entity"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB() {
	var err error

	host := os.Getenv("DATABASE_HOST")
	port := os.Getenv("DATABASE_PORT")
	user := os.Getenv("DATABASE_USER")
	pass := os.Getenv("DATABASE_PASSWORD")
	dbname := os.Getenv("DATABASE_NAME")

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable", host, user, pass, dbname, port)

	DB, err = gorm.Open(postgres.New(postgres.Config{
		DSN:                  dsn,
		PreferSimpleProtocol: true,
	}), &gorm.Config{})

	if err != nil {
		panic("failed to connect database")
	}

	err = DB.AutoMigrate(&entity.User{}, &entity.TopUpTransaction{}, &entity.Hotel{}, &entity.Room{}, &entity.Payment{}, &entity.Booking{})
	if err != nil {
		panic("failed to migrate database")
	}

	log.Println("Database connected")
}
