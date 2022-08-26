package entity

import (
	"time"

	"github.com/jinzhu/gorm"
)

type User struct {
	gorm.Model
	Name        string
	Email       string
	LastLogin   time.Time
	IsActive    bool
	IsStaff     bool
	IsSuperuser bool
}
