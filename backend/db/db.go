package db

import (
	"fmt"
	"os"

	"github.com/halmk/cliplist-ttv/backend/entity"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres" // Use PostgreSQL in gorm
)

var (
	db  *gorm.DB
	err error
)

// Init is initialize db from main function
func Init() {
	db_options := fmt.Sprintf(
		"host=%s port=%s user=%s dbname=%s password=%s sslmode=disable",
		os.Getenv("DB_HOST"), os.Getenv("DB_PORT"), os.Getenv("DB_USER"), os.Getenv("DB_NAME"), os.Getenv("DB_PASSWORD"),
	)
	db, err = gorm.Open("postgres", db_options)
	if err != nil {
		panic(err)
	}
	autoMigration()
}

// GetDB is called in models
func GetDB() *gorm.DB {
	return db
}

// Close is closing db
func Close() {
	if err := db.Close(); err != nil {
		panic(err)
	}
}

func autoMigration() {
	db.AutoMigrate(&entity.User{})
	db.AutoMigrate(&entity.Provider{})
	db.AutoMigrate(&entity.Socialaccount{})
	db.AutoMigrate(&entity.Socialtoken{})
}
