package repository

import (
	"fmt"
	"lux-hotel/entity"

	"gorm.io/gorm"
)

type HotelRepository interface {
	GetHotelList() ([]entity.Hotel, error)
	GetHotelDetail(id int) (entity.Hotel, error)
}

type hotelRepository struct {
	DB *gorm.DB
}

func NewHotelRepository(db *gorm.DB) HotelRepository {
	return &hotelRepository{DB: db}
}

func (hr *hotelRepository) GetHotelList() ([]entity.Hotel, error) {
	var hotels []entity.Hotel

	result := hr.DB.Preload("Rooms").Find(&hotels)

	if result.Error != nil {
		return nil, fmt.Errorf("500 | %v", result.Error)
	}

	return hotels, nil
}

func (hr *hotelRepository) GetHotelDetail(id int) (entity.Hotel, error) {
	var hotel entity.Hotel

	result := hr.DB.Preload("Rooms").First(&hotel, id)

	if result.Error != nil {
		if result.Error.Error() == "record not found" {
			return hotel, fmt.Errorf("404 | Hotel not found")
		}

		return hotel, fmt.Errorf("500 | %v", result.Error)
	}

	return hotel, nil
}
