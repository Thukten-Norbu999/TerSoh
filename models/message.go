package models

import (
    "time"
    "gorm.io/gorm"
)

type Message struct {
    gorm.Model
    Sender    string    ` + "`json:"sender"`" + `
    Recipient string    ` + "`json:"recipient"`" + `
    Content   string    ` + "`json:"content"`" + `
    Timestamp time.Time ` + "`json:"timestamp"`" + `
}
