package streamer

import (
	"github.com/halmk/cliplist-ttv/backend/db"
	"github.com/halmk/cliplist-ttv/backend/entity"
)

func Create(name string) (entity.Streamer, error) {
	db := db.GetDB()
	var streamer entity.Streamer
	{
		streamer.Name = name
	}
	if err := db.Create(&streamer).Error; err != nil {
		return streamer, err
	}

	return streamer, nil
}

func GetByName(name string) (entity.Streamer, error) {
	db := db.GetDB()
	var streamer entity.Streamer
	if err := db.Where("name = ?", name).First(&streamer).Error; err != nil {
		streamer, err = Create(name)
		if err != nil {
			return streamer, nil
		}
	}

	return streamer, nil
}

func GetByID(id int) (entity.Streamer, error) {
	db := db.GetDB()
	var streamer entity.Streamer
	if err := db.Where("id = ?", id).First(&streamer).Error; err != nil {
		return streamer, err
	}

	return streamer, nil
}
