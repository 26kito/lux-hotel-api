package config

import (
	"lux-hotel/middleware"
	"lux-hotel/repository"
	"lux-hotel/service"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

func Routes(DB *gorm.DB) {
	e := echo.New()

	userRepository := repository.NewUserRepository(DB)
	userService := service.NewUserService(userRepository)
	hotelRepository := repository.NewHotelRepository(DB)
	hotelService := service.NewHotelService(hotelRepository)
	midtransRepository := repository.NewMidtransRepository(DB)
	midtransService := service.NewMidtransService(midtransRepository)
	paymentRepository := repository.NewPaymentRepository(DB)
	paymentService := service.NewPaymentService(paymentRepository)

	api := e.Group("/api")

	// User
	api.POST("/users/register", userService.Register)
	api.POST("/users/login", userService.Login)
	api.GET("/users/balance", userService.GetBalance, middleware.ValidateJWTMiddleware)
	api.POST("/users/balance/top-up", userService.TopUpBalance, middleware.ValidateJWTMiddleware)

	// Hotel
	api.GET("/hotel-list", hotelService.GetHotelList)
	api.GET("/hotel/:id", hotelService.GetHotelDetail)
	api.POST("/hotel/:id/booking", hotelService.Booking, middleware.ValidateJWTMiddleware)

	// Payment
	api.POST("/order/payment", paymentService.Payment, middleware.ValidateJWTMiddleware)

	// Midtrans
	api.POST("/midtrans/callback", midtransService.HandleMidtransCallback)

	e.Logger.Fatal(e.Start(":8080"))
}
