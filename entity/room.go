package entity

type Room struct {
	ID         uint    `gorm:"primaryKey;autoIncrement"`
	HotelID    uint    `gorm:"not null" json:"hotel_id"`
	RoomNumber string  `gorm:"type:varchar(10);not null" json:"room_number"`
	RoomType   string  `gorm:"type:varchar(20);not null" json:"room_type"`
	Price      float64 `gorm:"type:decimal(10,2);not null" json:"price"`
	Status     string  `gorm:"type:varchar(10);not null" json:"status"`
}
