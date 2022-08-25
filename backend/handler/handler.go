package handler

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/halmk/cliplist-ttv/backend/utils/twitch"
	_ "github.com/heroku/x/hmetrics/onload"
	_ "github.com/lib/pq"
)

func Ping(c *gin.Context) {
	c.String(http.StatusOK, "pong")
}

func TwitchAPIRequest(c *gin.Context) {
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
	twitch := twitch.NewTwitchAppClient()
	response, status_code := twitch.GetRequest(api_url, params)
	c.JSON(status_code, gin.H{"response": response})
}

func TwitchLogin(c *gin.Context) {
	session := sessions.Default(c)
	c.SetSameSite(http.SameSiteNoneMode)

	email, ok := session.Get("loginUserEmail").(string)
	if ok && email != "" {
		log.Printf("User[%s] already logged in", email)
		c.Redirect(http.StatusFound, os.Getenv("LOGIN_REDIRECT_URL"))
		return
	}

	redirect_url, state, err := twitch.RedirectURL()
	if err != nil {
		c.String(http.StatusInternalServerError, "cannot get redirect url")
		return
	}
	c.SetCookie("oauth2_state", state, 3600, "/", os.Getenv("APP_DOMAIN"), true, true)
	c.Redirect(http.StatusFound, redirect_url)
}

func TwitchLoginCallback(c *gin.Context) {
	session := sessions.Default(c)

	code := c.Query("code")
	if code == "" {
		c.String(http.StatusBadRequest, "auth code doesn't exist")
		return
	}
	state := c.Query("state")
	cookie_state, err := c.Cookie("oauth2_state")
	if cookie_state == "" || err != nil {
		c.String(http.StatusBadRequest, "oauth2 state doesn't exist")
		return
	}
	if state != cookie_state {
		c.String(http.StatusBadRequest, "incorrect oauth2 state")
		return
	}

	// Exchange auth-code with access token
	tok, err := twitch.AccessToken(code)
	if err != nil {
		c.String(http.StatusInternalServerError, "cannot get access token")
		return
	}

	// Get user infomation
	twitch_client := twitch.NewTwitchClient(tok)
	info, status_code := twitch_client.GetUser()
	if status_code != 200 {
		c.String(http.StatusInternalServerError, "twitch request failed")
		return
	}

	// Update token of user's social account
	err = twitch.UpdateTokenInfo(info, tok)
	if err != nil {
		c.String(http.StatusInternalServerError, "user info update failed")
		return
	}

	email := info["email"].(string)
	session.Set("loginUserEmail", email)
	session.Options(sessions.Options{
		Path:     "/",
		Domain:   os.Getenv("APP_DOMAIN"),
		MaxAge:   60 * 60 * 24 * 30,
		SameSite: http.SameSiteNoneMode,
		Secure:   true,
		HttpOnly: false,
	})
	session.Save()

	log.Printf("User[%s] successful logged in", email)
	c.Redirect(http.StatusFound, os.Getenv("LOGIN_REDIRECT_URL"))
}

func Logout(c *gin.Context) {
	session := sessions.Default(c)
	email, _ := session.Get("loginUserEmail").(string)
	session.Clear()
	session.Options(sessions.Options{Path: "/", MaxAge: -1})
	session.Save()
	log.Printf("User[%s] logged out", email)
	c.Redirect(http.StatusFound, os.Getenv("LOGOUT_REDIRECT_URL"))
}
