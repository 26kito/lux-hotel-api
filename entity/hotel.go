package entity

type Hotel struct {
	ID            uint   `gorm:"primaryKey;autoIncrement"`
	Name          string `gorm:"type:varchar(100);not null" json:"name"`
	Location      string `gorm:"type:varchar(255);not null" json:"location"`
	ContactNumber string `gorm:"type:varchar(15)" json:"contact_number"`
	Email         string `gorm:"type:varchar(100)" json:"email"`
	Rooms         []Room `gorm:"foreignKey:HotelID" json:"rooms"`
}
