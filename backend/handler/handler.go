package handler

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"github.com/halmk/cliplist-ttv/backend/service/socialaccount"
	"github.com/halmk/cliplist-ttv/backend/service/socialtoken"
	"github.com/halmk/cliplist-ttv/backend/service/user"
	"github.com/halmk/cliplist-ttv/backend/utils/twitch"
	_ "github.com/heroku/x/hmetrics/onload"
	_ "github.com/lib/pq"
)

func Ping(c *gin.Context) {
	c.String(http.StatusOK, "pong")
}

func TwitchAPIAppRequest(c *gin.Context) {
	raw_query := c.Request.URL.RawQuery
	query := strings.Split(raw_query, "&")
	fmt.Println(query)
	req_url := MakeRequestURL(query)

	twitch := twitch.NewTwitchAppClient()
	response, status_code := twitch.GetRequest(req_url)
	c.JSON(status_code, gin.H{"response": response})
}

func TwitchAPIUserRequest(c *gin.Context) {
	bearer_token := c.Request.Header["Authorization"][0]
	token_string := strings.Split(bearer_token, " ")[1]
	token, ok := verifyJWT(token_string)
	if !ok {
		c.String(http.StatusBadRequest, "invalid token")
		return
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		c.String(http.StatusBadRequest, "invalid token claims")
		return
	}
	username := claims["username"].(string)

	raw_query := c.Request.URL.RawQuery
	query := strings.Split(raw_query, "?")
	fmt.Println(query)
	req_url := MakeRequestURL(query)

	user_record, err := user.GetByUsername(username)
	if err != nil {
		log.Println("Error(user.GetByUsername()): ", username)
		c.String(http.StatusInternalServerError, "could not get user by email")
		return
	}
	socialaccount_record, err := socialaccount.GetByUserId(user_record.ID)
	if err != nil {
		log.Println("Error(socialaccount.GetByUserId()): ", user_record.ID)
		c.String(http.StatusInternalServerError, "could not get socialaccount by user id")
		return
	}
	socialtoken_record, err := socialtoken.GetBySocialaccountId(socialaccount_record.ID)
	if err != nil {
		log.Println("Error(socialtoken.GetBySocialaccountId()): ", socialaccount_record.ID)
		c.String(http.StatusInternalServerError, "could not get socialtoken by socialaccount id")
		return
	}
	access_token := socialtoken_record.AccessToken

	twitch := twitch.NewTwitchUserClient(access_token)
	response, status_code := twitch.GetRequest(req_url)
	c.JSON(status_code, gin.H{"response": response})
}

func MakeRequestURL(query []string) string {
	params := make(map[string]string)
	api_url := ""

	for _, param := range query {
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

	req_url := api_url + "?"
	first := true
	for key, val := range params {
		if !first {
			req_url += "&"
		} else {
			first = false
		}
		req_url += key + "=" + val
	}
	return req_url
}

func TwitchLogin(c *gin.Context) {
	redirect_url, state, err := twitch.RedirectURL()
	if err != nil {
		c.String(http.StatusInternalServerError, "cannot get redirect url")
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"state":    state,
		"maxAge":   3600,
		"redirect": redirect_url,
	})
}

func TwitchLoginCallback(c *gin.Context) {
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
	twitch_client := twitch.NewTwitchUserClient(tok.AccessToken)
	info, status_code := twitch_client.GetRequest("https://api.twitch.tv/helix/users")
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

	email, ok := info["email"].(string)
	if !ok {
		c.String(http.StatusInternalServerError, "cannot get email")
		return
	}
	username, ok := info["login"].(string)
	if !ok {
		c.String(http.StatusInternalServerError, "cannot get username")
		return
	}

	expires := time.Now().Add(time.Hour * 24 * 30)
	claims := jwt.MapClaims{
		"username": username,
		"exp":      expires.Unix(),
	}

	// Generate JWT
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte(os.Getenv("JWT_SECRET")))

	log.Printf("User[%s(%s)] successful logged in", username, email)
	c.JSON(http.StatusOK, gin.H{
		"token":   tokenString,
		"expires": expires,
	})
}

func verifyJWT(tokenString string) (*jwt.Token, bool) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(os.Getenv("JWT_SECRET")), nil
	})
	if err != nil || !token.Valid {
		log.Println(token, err)
		return nil, false
	}
	return token, true
}

func Logout(c *gin.Context) {
	bearer_token := c.Request.Header["Authorization"][0]
	token_string := strings.Split(bearer_token, " ")[1]

	token, ok := verifyJWT(token_string)
	if !ok {
		log.Println("invalid token detected")
	}
	claims := token.Claims.(jwt.MapClaims)
	username := claims["username"].(string)

	log.Printf("User[%s] logged out", username)
	c.String(http.StatusOK, "logged out")
}
