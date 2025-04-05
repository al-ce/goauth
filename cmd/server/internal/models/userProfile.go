package models

import "time"

type UserProfile struct {
	Email     string     `gorm:"type:varchar(255);not null;unique"`
	LastLogin *time.Time `gorm:"type:timestamp"`
}
