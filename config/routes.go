package config

import (
	"lux-hotel/repository"
	"lux-hotel/service"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

func Routes(DB *gorm.DB) {
	e := echo.New()

	userRepository := repository.NewUserRepository(DB)
	userService := service.NewUserService(userRepository)

	api := e.Group("/api")

	// User routes
	api.POST("/users/register", userService.Register)

	e.Logger.Fatal(e.Start(":8080"))
}
