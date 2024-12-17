package main

import (
	"lux-hotel/config"

	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()

	config.InitDB()
	DB := config.DB

	config.Routes(DB)
}
