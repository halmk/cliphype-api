package entity

import (
	"github.com/jinzhu/gorm"
)

type Streamer struct {
	gorm.Model
	Name string
}
