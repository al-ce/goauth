package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID                  uuid.UUID  `gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	Email               string     `gorm:"type:varchar(255);not null;unique"`
	Password            string     `gorm:"type:text;not null"`
	LastLogin           *time.Time `gorm:"type:timestamp"`
	FailedLoginAttempts int        `gorm:"type:integer;default:0"`
	AccountLocked       bool       `gorm:"type:boolean;default:false"`
	AccountLockedUntil  *time.Time `gorm:"type:timestamp"`
}
