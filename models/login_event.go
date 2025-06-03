package models

import (
    "time"
    "gorm.io/gorm"
)

type LoginEvent struct {
    gorm.Model
    Username  string    ` + "`gorm:"index;not null"`" + `
    Timestamp time.Time ` + "`json:"autoCreateTime"`" + `
}
