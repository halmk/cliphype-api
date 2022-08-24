package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/halmk/cliplist-ttv/backend/db"
	"github.com/halmk/cliplist-ttv/backend/server"
	_ "github.com/heroku/x/hmetrics/onload"
	_ "github.com/lib/pq"
)

func main() {
	CheckEnvs()

	db.Init()
	defer db.Close()

	port := os.Getenv("PORT")
	r := server.SetupRouter()
	r.Run(":" + port)
}

func CheckEnvs() {
	envs := [...]string{
		"PORT",
		"TWITCH_CLIENT_ID",
		"TWITCH_CLIENT_SECRET",
		"APP_ORIGIN",
		"APP_DOMAIN",
		"BASE_URL",
		"LOGIN_REDIRECT_URL",
		"LOGOUT_REDIRECT_URL",
		"SESSION_SECRET",
		"DB_HOST",
		"DB_NAME",
		"DB_USER",
		"DB_PASSWORD",
		"DB_PORT",
	}

	for _, env := range envs {
		CheckEnv(env)
	}
}

func CheckEnv(env string) {
	val := os.Getenv(env)
	if val == "" {
		log.Fatal(fmt.Sprintf("$%s must be set", strings.ToUpper(env)))
	}
}
