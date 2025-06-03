package models

import (
    "time"
    "gorm.io/gorm"
)

type Post struct {
    gorm.Model
    Currency  string    ` + "`gorm:"not null" json:"currency"`" + `
    Rate      float64   ` + "`json:"rate"`" + `
    CreatedAt time.Time ` + "`json:"created_at"`" + `
}
