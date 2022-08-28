package entity

import (
	"github.com/jinzhu/gorm"
)

type Provider struct {
	gorm.Model
	Name string
}
