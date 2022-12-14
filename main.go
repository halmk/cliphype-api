package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/halmk/cliphype-api/db"
	"github.com/halmk/cliphype-api/estimator"
	"github.com/halmk/cliphype-api/server"
	_ "github.com/heroku/x/hmetrics/onload"
	_ "github.com/lib/pq"
)

func main() {
	CheckEnvs()

	db.Init()
	defer db.Close()

	estimator.Init()

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
		"API_DOMAIN",
		"APP_BASE_URL",
		"LOGIN_REDIRECT_URL",
		"LOGOUT_REDIRECT_URL",
		"SESSION_SECRET",
		"JWT_SECRET",
		"DATABASE_URL",
		"AWS_REGION",
		"AWS_ACCESS_KEY_ID",
		"AWS_SECRET_ACCESS_KEY",
		"S3_BUCKET_NAME",
		"S3_MODEL_KEY",
	}

	for _, env := range envs {
		CheckEnv(env)
	}
}

func CheckEnv(env string) {
	val := os.Getenv(env)
	if val == "" {
		log.Fatalf(fmt.Sprintf("$%s must be set", strings.ToUpper(env)))
	}
}
