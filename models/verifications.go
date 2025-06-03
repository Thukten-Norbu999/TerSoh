package models

import "gorm.io/gorm"

type Verification struct {
	gorm.Model
	User_Idf   string `gorm:"not null" json:"user_idf"`
	Session_Id string `gorm:"unqiue" json:"session_id"`
	Status     string `gorm:"not null"`

	User User `gorm:"foreignKey:User_Idf;references:Identification_No" json:"user"`
}
