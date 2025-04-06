package models

import (
    "time"
    "github.com/google/uuid"
)

type Session struct {
    ID        uuid.UUID  `gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
    UserID    uuid.UUID  `gorm:"type:uuid;not null;index"`
    Token     string     `gorm:"type:text;not null;unique"`
    ExpiresAt time.Time  `gorm:"type:timestamp;not null"`
    CreatedAt time.Time  `gorm:"type:timestamp;not null;default:now()"`
    IP        string     `gorm:"type:varchar(45)"`
    UserAgent string     `gorm:"type:text"`
}
