package main

import (
	"lux-hotel/config"

	"github.com/joho/godotenv"
)

// @title API Documentation
// @version 1.0
// @description This is the API documentation for Lux Hotel application
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
func main() {
	godotenv.Load()

	config.InitDB()
	DB := config.DB

	config.Routes(DB)
}
