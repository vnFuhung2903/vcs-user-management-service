package entities

type UserScope struct {
	ID   string `gorm:"primaryKey"`
	Name string `gorm:"type:varchar(50);unique;not null"`
}
