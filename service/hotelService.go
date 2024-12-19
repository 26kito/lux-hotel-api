package service

import (
	"fmt"
	"lux-hotel/entity"
	"lux-hotel/repository"
	"strconv"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

type HotelService interface {
	GetHotelList(c echo.Context) error
	GetHotelDetail(c echo.Context) error
	Booking(c echo.Context) error
}

type hotelService struct {
	HotelRepository repository.HotelRepository
}

func NewHotelService(hotelRepository repository.HotelRepository) HotelService {
	return &hotelService{HotelRepository: hotelRepository}
}

func (hs *hotelService) GetHotelList(c echo.Context) error {
	hotels, err := hs.HotelRepository.GetHotelList()

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
		Message: "Success",
		Data:    hotels,
	})
}

func (hs *hotelService) GetHotelDetail(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))

	if err != nil {
		return c.JSON(400, entity.ResponseError{
			Status:  400,
			Message: "Invalid ID",
		})
	}

	hotel, err := hs.HotelRepository.GetHotelDetail(id)

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
		Message: "Success",
		Data:    hotel,
	})
}

func (hs *hotelService) Booking(c echo.Context) error {
	userID := c.Get("user").(jwt.MapClaims)["user_id"].(float64)
	hotelID, err := strconv.Atoi(c.Param("id"))

	if err != nil {
		return c.JSON(400, entity.ResponseError{
			Status:  400,
			Message: "Invalid ID",
		})
	}

	var payload entity.BookingRequest
	if err := c.Bind(&payload); err != nil {
		return c.JSON(400, entity.ResponseError{
			Status:  400,
			Message: "Invalid request",
		})
	}

	if err := validateBookingPayload(payload); err != nil {
		errCode, _ := strconv.Atoi(err.Error()[:3])
		errMessage := err.Error()[6:]

		return c.JSON(errCode, entity.ResponseError{
			Status:  errCode,
			Message: errMessage,
		})
	}

	response, err := hs.HotelRepository.Booking(int(userID), hotelID, payload)

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
		Message: "Success",
		Data: map[string]interface{}{
			"order_id":     response.OrderID,
			"booking_code": response.BookingCode,
			"check_in":     response.CheckIn,
			"check_out":    response.CheckOut,
			"total_price":  response.TotalPrice,
		},
	})
}

func validateBookingPayload(payload entity.BookingRequest) error {
	if payload.RoomID == 0 {
		return fmt.Errorf("400 | room ID is required")
	}

	if payload.CheckIn == "" {
		return fmt.Errorf("400 | check in date is required")
	}

	if payload.CheckOut == "" {
		return fmt.Errorf("400 | check out date is required")
	}

	return nil
}
