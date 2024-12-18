package service

import (
	"lux-hotel/entity"
	"lux-hotel/repository"
	"strconv"

	"github.com/labstack/echo/v4"
)

type HotelService interface {
	GetHotelList(c echo.Context) error
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
