package entity

type Booking struct {
	ID         uint    `gorm:"primaryKey;autoIncrement"`
	GuestID    uint    `gorm:"not null" json:"guest_id"`
	RoomID     uint    `gorm:"not null" json:"room_id"`
	CheckIn    string  `gorm:"type:date;not null" json:"check_in"`
	CheckOut   string  `gorm:"type:date;not null" json:"check_out"`
	TotalDays  int     `gorm:"not null" json:"total_days"`
	TotalPrice float64 `gorm:"type:decimal(10,2);not null" json:"total_price"`
	Status     string  `gorm:"type:varchar(10);not null" json:"status"`
}
