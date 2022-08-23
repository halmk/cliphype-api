package server

import (
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/halmk/cliplist-ttv/backend/handler"
	_ "github.com/heroku/x/hmetrics/onload"
	_ "github.com/lib/pq"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins: []string{
			os.Getenv("APP_ORIGIN"),
		},
		AllowMethods: []string{
			"POST",
			"GET",
			"OPTIONS",
		},
		AllowHeaders: []string{
			"Access-Control-Allow-Credentials",
			"Access-Control-Allow-Headers",
			"Content-Type",
			"Content-Length",
			"Accept-Encoding",
			"Authorization",
		},
		AllowCredentials: true,
		MaxAge:           24 * time.Hour,
	}))

	session_secret := os.Getenv("SESSION_SECRET")
	store := cookie.NewStore([]byte(session_secret))
	r.Use(sessions.Sessions("session", store))

	// Ping test
	r.GET("/ping", handler.Ping)

	// API
	api := r.Group("/api")
	{
		// Proxy Twitch API Request
		api.GET("/twitch", handler.TwitchAPIRequest)
	}

	// Accounts
	accounts := r.Group("/accounts")
	{
		twitch := accounts.Group("/twitch")
		{
			twitch.GET("/login", handler.TwitchLogin)
			twitch.GET("/login/callback", handler.TwitchLoginCallback)
			twitch.GET("/logout", handler.TwitchLogout)
		}
	}

	return r
}
