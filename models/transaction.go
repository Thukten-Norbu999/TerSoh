package models

import (
    "gorm.io/gorm"
)

type Transaction struct {
    gorm.Model
    Amount   float64 ` + "`json:"amount"`" + `
    Currency string  ` + "`json:"currency"`" + `
}
