package provider

import (
	"github.com/halmk/cliphype-api/db"
	"github.com/halmk/cliphype-api/entity"
)

func Create(name string) (entity.Provider, error) {
	db := db.GetDB()
	var p entity.Provider
	p.Name = name
	if err := db.Create(&p).Error; err != nil {
		return p, err
	}

	return p, nil
}

func GetByName(name string) (entity.Provider, error) {
	db := db.GetDB()
	var p entity.Provider
	if err := db.Where("name = ?", name).First(&p).Error; err != nil {
		return p, err
	}

	return p, nil
}
