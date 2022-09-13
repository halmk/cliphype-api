package handler

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"github.com/halmk/cliplist-ttv/backend/entity"
	"github.com/halmk/cliplist-ttv/backend/estimator"
	"github.com/halmk/cliplist-ttv/backend/service/playlist"
	"github.com/halmk/cliplist-ttv/backend/service/playlistclip"
	"github.com/halmk/cliplist-ttv/backend/service/socialaccount"
	"github.com/halmk/cliplist-ttv/backend/service/socialtoken"
	"github.com/halmk/cliplist-ttv/backend/service/streamer"
	"github.com/halmk/cliplist-ttv/backend/service/user"
	"github.com/halmk/cliplist-ttv/backend/utils/twitch"
	_ "github.com/heroku/x/hmetrics/onload"
	_ "github.com/lib/pq"
	"github.com/sajari/word2vec"
)

func Ping(c *gin.Context) {
	c.String(http.StatusOK, "pong")
}

func TwitchAPIAppRequest(c *gin.Context) {
	raw_query := c.Request.URL.RawQuery
	query, _ := url.QueryUnescape(raw_query)
	log.Println("Query app requested:", query)
	req_url := MakeRequestURL(raw_query)

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
	query, _ := url.QueryUnescape(raw_query)
	log.Println("Query user requested:", query)
	req_url := MakeRequestURL(raw_query)

	// Get user's access token
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
	refresh_token := socialtoken_record.RefreshToken

	twitch := twitch.NewTwitchUserClient(username, access_token, refresh_token)
	response, status_code := twitch.GetRequest(req_url)
	c.JSON(status_code, gin.H{"response": response})
}

func MakeRequestURL(query string) string {
	param_map := make(map[string][]string)
	api_url := ""

	params := strings.Split(query, "&")
	for _, param := range params {
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
				param_map[key] = append(param_map[key], value)
			}
		}
	}

	req_url := api_url + "?"
	first := true
	for key, vals := range param_map {
		for _, val := range vals {
			if !first {
				req_url += "&"
			} else {
				first = false
			}
			req_url += key + "=" + val
		}
	}
	return req_url
}

func TwitchLogin(c *gin.Context) {
	redirect_url, state, err := twitch.RedirectURL()
	if err != nil {
		c.String(http.StatusInternalServerError, "cannot get redirect url")
		return
	}
	c.SetSameSite(http.SameSiteNoneMode)
	c.SetCookie("oauth2_state", state, 3600, "/", os.Getenv("API_DOMAIN"), true, false)
	c.Redirect(http.StatusFound, redirect_url)
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
	twitch_client := twitch.NewTwitchUserClient("", tok.AccessToken, tok.RefreshToken)
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

type PlaylistResponse struct {
	Playlist Playlist `json:"playlist"`
	Clips    []Clip   `json:"clips"`
}

type Playlist struct {
	ID        uint      `json:"id"`
	Title     string    `json:"title"`
	Streamer  string    `json:"streamer"`
	Creator   string    `json:"creator"`
	CreatedAt time.Time `json:"createdAt"`
}

type Clip struct {
	ID           string  `json:"id"`
	Duration     float64 `json:"duration"`
	EmbedURL     string  `json:"embed_url"`
	ThumbnailURL string  `json:"thumbnail_url"`
	Title        string  `json:"title"`
	URL          string  `json:"url"`
	VideoID      string  `json:"video_id"`
	VodOffset    int     `json:"vod_offset"`
}

func GetVideoRange(created_at string, duration string) (string, string) {
	started_at := ""
	ended_at := ""

	split_duration := strings.Split(duration, "h")
	hours := 0
	minutes := 0
	seconds := 0
	if len(split_duration) == 2 {
		hours, _ = strconv.Atoi(split_duration[0])
	}
	duration = split_duration[len(split_duration)-1]
	split_duration = strings.Split(duration, "m")
	if len(split_duration) == 2 {
		minutes, _ = strconv.Atoi(split_duration[0])
	}
	duration = split_duration[len(split_duration)-1]
	split_duration = strings.Split(duration, "s")
	if len(split_duration) == 2 {
		seconds, _ = strconv.Atoi(split_duration[0])
	}
	time_hours := time.Duration(hours) * time.Hour
	time_minutes := time.Duration(minutes) * time.Minute
	time_seconds := time.Duration(seconds) * time.Second

	st, _ := time.Parse(time.RFC3339, created_at)
	et := st.Add(time_hours + time_minutes + time_seconds)

	started_at = st.Format(time.RFC3339)
	ended_at = et.Format(time.RFC3339)

	return started_at, ended_at
}

func SortClips(clips []map[string]interface{}) []map[string]interface{} {
	sort.Slice(clips, func(i, j int) bool {
		return clips[i]["created_at"].(string) < clips[j]["created_at"].(string)
	})
	return clips
}

func CreatePlaylistForLatestVideo(c *gin.Context, streamer entity.Streamer, streamer_name string) {
	twitch := twitch.NewTwitchAppClient()
	user_body, _ := twitch.GetUser(streamer_name)
	user_entity := user_body["data"].([]interface{})[0].(map[string]interface{})
	first := 1
	video_body, _ := twitch.GetVideos(user_entity["id"].(string), &first)
	video := video_body["data"].([]interface{})[0].(map[string]interface{})
	video_id := video["id"].(string)
	video_thumbnail := video["thumbnail_url"].(string)
	if !strings.Contains(video_thumbnail, "thumb") {
		return
	}
	if p, err := playlist.GetByVideoID(video_id); len(p) == 0 && err == nil {
		log.Println("Generate new playlist for the latest video")
		started_at, ended_at := GetVideoRange(video["created_at"].(string), video["duration"].(string))
		first = 10
		clips_body, _ := twitch.GetClips(user_entity["id"].(string), &first, &started_at, &ended_at)
		clips := clips_body["data"].([]interface{})
		var clips_map []map[string]interface{}
		for _, clip := range clips {
			clips_map = append(clips_map, clip.(map[string]interface{}))
		}
		sorted_clips := SortClips(clips_map)
		title := fmt.Sprintf("Most viewed clips for the latest video - %s", video_id)
		if len(sorted_clips) > 1 {
			playlist_entity, _ := playlist.Create(title, streamer, nil, &video_id)
			for _, clip := range sorted_clips {
				_, err := playlistclip.Create(clip["id"].(string), clip["duration"].(float64), clip["embed_url"].(string), clip["thumbnail_url"].(string), clip["title"].(string), clip["url"].(string), clip["video_id"].(string), int(clip["vod_offset"].(float64)), playlist_entity)
				if err != nil {
					c.String(http.StatusInternalServerError, "Failed creating playlists")
					return
				}
			}
		}
	}
}

func GetPlaylists(c *gin.Context) {
	// Analyse params
	streamer_name := c.Query("streamer")
	streamer, _ := streamer.GetByName(streamer_name)

	// Generate a playlist for the latest video
	CreatePlaylistForLatestVideo(c, streamer, streamer_name)

	playlists, err := playlist.GetByStreamerID(streamer.ID)
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed getting playlists")
		return
	}

	var pr []PlaylistResponse
	for _, playlist := range playlists {
		clips, err := playlistclip.GetArrayByPlaylist(playlist.ID)
		if err != nil {
			c.String(http.StatusInternalServerError, "Failed getting playlist clips")
			return
		}
		creator, _ := user.GetByID(playlist.CreatorID)
		var cs []Clip
		for _, clip := range clips {
			c := Clip{
				ID:           clip.ClipID,
				Duration:     clip.Duration,
				EmbedURL:     clip.EmbedURL,
				ThumbnailURL: clip.ThumbnailURL,
				Title:        clip.Title,
				URL:          clip.URL,
				VideoID:      clip.VideoID,
				VodOffset:    clip.VodOffset,
			}
			cs = append(cs, c)
		}
		p := PlaylistResponse{
			Playlist: Playlist{
				ID:        playlist.ID,
				Title:     playlist.Title,
				Streamer:  streamer_name,
				Creator:   creator.Name,
				CreatedAt: playlist.CreatedAt,
			},
			Clips: cs,
		}
		pr = append(pr, p)
	}

	c.JSON(http.StatusOK, pr)
}

type PlaylistParams struct {
	Streamer string       `json:"streamer"`
	Creator  string       `json:"creator"`
	Title    string       `json:"title"`
	Clips    []ClipParams `json:"clips"`
}

type ClipParams struct {
	ID           string  `json:"id"`
	Duration     float64 `json:"duration"`
	EmbedURL     string  `json:"embed_url"`
	ThumbnailURL string  `json:"thumbnail_url"`
	Title        string  `json:"title"`
	URL          string  `json:"url"`
	VideoID      string  `json:"video_id"`
	VodOffset    int     `json:"vod_offset"`
}

func PostPlaylists(c *gin.Context) {
	var pp PlaylistParams
	if err := c.BindJSON(&pp); err != nil {
		c.String(http.StatusBadRequest, "Failed binding request parameters")
		return
	}
	streamer, err := streamer.GetByName(pp.Streamer)
	if err != nil {
		c.String(http.StatusBadRequest, "Failed getting streamer")
		return
	}
	creator, err := user.GetByUsername(pp.Creator)
	if err != nil {
		c.String(http.StatusBadRequest, "Failed getting creator")
		return
	}

	playlist, err := playlist.Create(pp.Title, streamer, &creator, nil)
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed creating playlist")
		return
	}

	for _, clip := range pp.Clips {
		_, err := playlistclip.Create(clip.ID, clip.Duration, clip.EmbedURL, clip.ThumbnailURL, clip.Title, clip.URL, clip.VideoID, clip.VodOffset, playlist)
		if err != nil {
			c.String(http.StatusInternalServerError, "Failed creating playlist clip")
			return
		}
	}

	c.String(http.StatusOK, "successful post a playlist")
}

func GetChatbot(c *gin.Context) {
	chatbot_name := os.Getenv("CHATBOT")
	chatbot_user, err := user.GetByUsername(chatbot_name)
	if err != nil {
		c.String(http.StatusInternalServerError, "chatbot doesn't exist")
		return
	}
	chatbot_socialaccount, err := socialaccount.GetByUserId(chatbot_user.ID)
	if err != nil {
		c.String(http.StatusInternalServerError, "chatbot doesn't exist")
		return
	}
	chatbot_socialtoken, err := socialtoken.GetBySocialaccountId(chatbot_socialaccount.ID)
	if err != nil {
		c.String(http.StatusInternalServerError, "chatbot doesn't exist")
		return
	}

	twitch_client := twitch.NewTwitchUserClient("", chatbot_socialtoken.AccessToken, chatbot_socialtoken.RefreshToken)
	_, status_code := twitch_client.GetRequest("https://api.twitch.tv/helix/users")
	if status_code != 200 {
		c.String(http.StatusInternalServerError, "twitch request failed")
		return
	}

	access_token := twitch_client.AccessToken
	c.JSON(http.StatusOK, gin.H{
		"username": chatbot_name,
		"password": access_token,
	})
}

func GetHypes(c *gin.Context) {
	sentence := c.Query("sentence")
	if len(sentence) == 0 {
		c.String(http.StatusBadRequest, "sentence param required")
		return
	}

	var hypes []map[string]interface{}
	model := estimator.GetEstimator()
	hypewords := estimator.GetHypewords()

	expr_message := word2vec.Expr{}
	expr_message.Add(1, sentence)
	_, err := expr_message.Eval(model)
	log.Println(sentence, err)

	for _, hypeword := range *hypewords {
		hype := make(map[string]interface{})
		hype["word"] = hypeword

		expr_hypeword := word2vec.Expr{}
		expr_hypeword.Add(1, hypeword)
		pair := [2]word2vec.Expr{
			expr_message,
			expr_hypeword,
		}
		pairs := [][2]word2vec.Expr{pair}
		value, err := model.Coses(pairs)
		if err != nil {
			hype["value"] = 0.0
		} else {
			hype["value"] = value
		}
		hypes = append(hypes, hype)
	}

	c.JSON(http.StatusOK, gin.H{
		"sentence": sentence,
		"hypes":    hypes,
	})
}
