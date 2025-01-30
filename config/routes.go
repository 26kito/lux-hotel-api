package config

import (
	customeMiddleware "lux-hotel/middleware"
	"lux-hotel/repository"
	"lux-hotel/service"
	"net/http"

	_ "lux-hotel/docs"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	echoSwagger "github.com/swaggo/echo-swagger"
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

	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"http://localhost:5173"},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept},
		AllowMethods: []string{http.MethodGet, http.MethodHead, http.MethodPut, http.MethodPatch, http.MethodPost, http.MethodDelete},
	}))

	api := e.Group("/api")

	// User
	api.POST("/users/register", userService.Register)
	api.POST("/users/login", userService.Login)
	api.POST("/users/check-email", userService.GetUserByEmail)
	api.GET("/users/balance", userService.GetBalance, customeMiddleware.ValidateJWTMiddleware)
	api.POST("/users/balance/top-up", userService.TopUpBalance, customeMiddleware.ValidateJWTMiddleware)
	api.GET("/users/book/history", userService.GetBookHistory, customeMiddleware.ValidateJWTMiddleware)

	// Hotel
	api.GET("/hotel-list", hotelService.GetHotelList)
	api.GET("/hotel/:id", hotelService.GetHotelDetail)
	api.POST("/hotel/:id/booking", hotelService.Booking, customeMiddleware.ValidateJWTMiddleware)

	// Payment
	api.POST("/order/payment", paymentService.Payment, customeMiddleware.ValidateJWTMiddleware)

	// Midtrans Callback
	api.POST("/midtrans/callback", midtransService.HandleMidtransCallback)

	api.GET("/swagger/*", echoSwagger.WrapHandler)

	e.Logger.Fatal(e.Start(":8080"))
}
