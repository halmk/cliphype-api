package main

import (
	"log"
	"os"

	"github.com/halmk/cliplist-ttv/backend/server"
	_ "github.com/heroku/x/hmetrics/onload"
	_ "github.com/lib/pq"
)

func main() {
	CheckEnv()
	port := os.Getenv("PORT")
	r := server.SetupRouter()
	r.Run(":" + port)
}

func CheckEnv() {
	port := os.Getenv("PORT")
	twitch_client_id := os.Getenv("TWITCH_CLIENT_ID")
	twitch_client_secret := os.Getenv("TWITCH_CLIENT_SECRET")
	app_origin := os.Getenv("APP_ORIGIN")
	base_url := os.Getenv("BASE_URL")

	if port == "" {
		log.Fatal("$PORT must be set")
	}
	if twitch_client_id == "" {
		log.Fatal("$TWITCH_CLIENT_ID must be set")
	}
	if twitch_client_secret == "" {
		log.Fatal("$TWITCH_CLIENT_SECRET must be set")
	}
	if app_origin == "" {
		log.Fatal("$APP_ORIGIN must be set")
	}
	if base_url == "" {
		log.Fatal("$BASE_URL must be set")
	}
}
