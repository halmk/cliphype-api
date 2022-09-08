package server

import (
	"os"
	"time"

	"github.com/gin-contrib/cors"
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

	// Ping test
	r.GET("/ping", handler.Ping)

	// API
	api := r.Group("/api")
	{
		// Proxy Twitch API Request
		twitch := api.Group("/twitch")
		{
			twitch.GET("/app", handler.TwitchAPIAppRequest)
			twitch.GET("/user", handler.TwitchAPIUserRequest)
		}

		// clip playlists API Request
		api.GET("/playlists", handler.GetPlaylists)
		api.POST("/playlists", handler.PostPlaylists)

		// Get chatbot infomation
		api.GET("/chatbot", handler.GetChatbot)
	}

	// Accounts
	accounts := r.Group("/accounts")
	{
		accounts.GET("/logout", handler.Logout)
		twitch := accounts.Group("/twitch")
		{
			twitch.GET("/login", handler.TwitchLogin)
			twitch.GET("/login/callback", handler.TwitchLoginCallback)
		}
	}

	return r
}
