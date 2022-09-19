package db

import (
	"os"

	"github.com/halmk/cliphype-api/entity"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres" // Use PostgreSQL in gorm
)

var (
	db  *gorm.DB
	err error
)

// Init is initialize db from main function
func Init() {
	db_url := os.Getenv("DATABASE_URL")
	db, err = gorm.Open("postgres", db_url)
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
	db.AutoMigrate(&entity.Playlist{})
	db.AutoMigrate(&entity.PlaylistClip{})
	db.AutoMigrate(&entity.Streamer{})
	db.AutoMigrate(&entity.AutoClip{})
}
