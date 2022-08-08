package main

import (
	"backend/twitch_api"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "github.com/heroku/x/hmetrics/onload"
	_ "github.com/lib/pq"
)

func dbFunc(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if _, err := db.Exec("CREATE TABLE IF NOT EXISTS ticks (tick timestamp)"); err != nil {
			c.String(http.StatusInternalServerError,
				fmt.Sprintf("Error creating database table: %q", err))
			return
		}

		if _, err := db.Exec("INSERT INTO ticks VALUES (now())"); err != nil {
			c.String(http.StatusInternalServerError,
				fmt.Sprintf("Error incrementing tick: %q", err))
			return
		}

		rows, err := db.Query("SELECT tick FROM ticks")
		if err != nil {
			c.String(http.StatusInternalServerError,
				fmt.Sprintf("Error reading ticks: %q", err))
			return
		}

		defer rows.Close()
		for rows.Next() {
			var tick time.Time
			if err := rows.Scan(&tick); err != nil {
				c.String(http.StatusInternalServerError,
					fmt.Sprintf("Error scanning ticks: %q", err))
				return
			}
			c.String(http.StatusOK, fmt.Sprintf("Read from DB: %s\n", tick.String()))
		}
	}
}

func setupRouter() *gin.Engine {
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins: []string{
			"http://localhost:3000",
			"https://cliphype.netlify.app",
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
	r.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("Error opening database: %q", err)
	}
	r.GET("/db", dbFunc(db))

	api := r.Group("/api")
	{
		api.GET("/twitch", func(c *gin.Context) {
			raw_query := c.Request.URL.RawQuery
			query_arr := strings.Split(raw_query, "&")
			fmt.Println(query_arr)
			params := make(map[string]string)
			api_url := ""

			for _, param := range query_arr {
				tStr := strings.Split(param, "=")
				key := tStr[0]
				value := tStr[1]
				if key == "url" {
					parsed_url, err := url.QueryUnescape(value)
					if err != nil {
						fmt.Println(err)
					}
					api_url = parsed_url
				} else {
					if len(value) != 0 {
						params[key] = value
					}
				}
			}
			fmt.Println(api_url, params)
			twitch := twitch_api.NewTwitchAPI()
			response, status_code := twitch.GetRequest(api_url, params)
			c.JSON(status_code, gin.H{"response": response})
		})
	}

	return r
}

func main() {
	port := os.Getenv("PORT")
	twitch_client_id := os.Getenv("TWITCH_CLIENT_ID")
	twitch_client_secret := os.Getenv("TWITCH_CLIENT_SECRET")

	if port == "" {
		log.Fatal("$PORT must be set")
	}
	if twitch_client_id == "" {
		log.Fatal("$TWITCH_CLIENT_ID must be set")
	}
	if twitch_client_secret == "" {
		log.Fatal("$TWITCH_CLIENT_SECRET must be set")
	}

	r := setupRouter()
	r.Run(":" + port)
}
