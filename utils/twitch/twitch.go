package twitch

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/halmk/cliphype-api/db"
	"github.com/halmk/cliphype-api/service/provider"
	"github.com/halmk/cliphype-api/service/socialaccount"
	"github.com/halmk/cliphype-api/service/socialtoken"
	"github.com/halmk/cliphype-api/service/user"
	"golang.org/x/oauth2"
)

type TwitchAppClient struct {
	client_id     string
	client_secret string
	token         string
	count         int
}

type TwitchUserClient struct {
	Username     string
	ClientID     string
	ClientSecret string
	AccessToken  string
	RefreshToken string
	Count        int
}

func NewTwitchAppClient() TwitchAppClient {
	client_id := os.Getenv("TWITCH_CLIENT_ID")
	client_secret := os.Getenv("TWITCH_CLIENT_SECRET")

	twitch := TwitchAppClient{
		client_id,
		client_secret,
		"",
		0,
	}
	err := twitch.ReadTokenFile()
	if err != nil {
		fmt.Println(err)
	}
	return twitch
}

func NewTwitchUserClient(username, access_token, refresh_token string) TwitchUserClient {
	client_id := os.Getenv("TWITCH_CLIENT_ID")
	client_secret := os.Getenv("TWITCH_CLIENT_SECRET")
	twitch := TwitchUserClient{
		username,
		client_id,
		client_secret,
		access_token,
		refresh_token,
		0,
	}
	return twitch
}

func (twitch *TwitchAppClient) GetToken() {
	req_url := "https://id.twitch.tv/oauth2/token"

	values := url.Values{}
	values.Set("grant_type", "client_credentials")
	values.Add("client_id", twitch.client_id)
	values.Add("client_secret", twitch.client_secret)

	req, err := http.NewRequest(
		"POST",
		req_url,
		strings.NewReader(values.Encode()),
	)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := new(http.Client)
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Fatal(resp.Status)
	}

	byteArray, _ := ioutil.ReadAll(resp.Body)
	var mapBody map[string]interface{}
	json.Unmarshal(byteArray, &mapBody)
	twitch.token = mapBody["access_token"].(string)
	twitch.WriteTokenFile()
}

func (twitch *TwitchAppClient) GetRequest(url string) (map[string]interface{}, int) {
	twitch.count++
	req, _ := http.NewRequest(
		"GET",
		url,
		nil,
	)
	req.Header.Add("Authorization", "Bearer "+twitch.token)
	req.Header.Add("Client-ID", twitch.client_id)
	log.Println(req.URL)

	client := new(http.Client)
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Println(resp.Status)
		if twitch.count <= 1 {
			twitch.GetToken()
			return twitch.GetRequest(url)
		} else {
			return nil, resp.StatusCode
		}
	} else {
		byteArray, _ := ioutil.ReadAll(resp.Body)
		var mapBody map[string]interface{}
		json.Unmarshal(byteArray, &mapBody)
		return mapBody, resp.StatusCode
	}
}

func (twitch *TwitchAppClient) GetUser(login string) (map[string]interface{}, int) {
	req_url := fmt.Sprintf("https://api.twitch.tv/helix/users?login=%s", login)
	return twitch.GetRequest(req_url)
}

func (twitch *TwitchAppClient) GetVideos(user_id string, first *int) (map[string]interface{}, int) {
	req_url := fmt.Sprintf("https://api.twitch.tv/helix/videos?user_id=%s", user_id)
	if first != nil {
		req_url += "&first=" + strconv.Itoa(*first)
	}
	return twitch.GetRequest(req_url)
}

func (twitch *TwitchAppClient) GetClips(broadcaster_id string, first *int, started_at *string, ended_at *string) (map[string]interface{}, int) {
	req_url := fmt.Sprintf("https://api.twitch.tv/helix/clips?broadcaster_id=%s", broadcaster_id)
	if first != nil {
		req_url += "&first=" + strconv.Itoa(*first)
	}
	if started_at != nil {
		req_url += "&started_at=" + *started_at
	}
	if ended_at != nil {
		req_url += "&ended_at=" + *ended_at
	}
	return twitch.GetRequest(req_url)
}

func (twitch *TwitchAppClient) ReadTokenFile() error {
	b, err := ioutil.ReadFile("twitch_app_token.txt")
	if err != nil {
		twitch.GetToken()
		return twitch.WriteTokenFile()
	}
	twitch.token = string(b)
	return nil
}

func (twitch *TwitchAppClient) WriteTokenFile() error {
	err := ioutil.WriteFile("twitch_app_token.txt", []byte(twitch.token), 0666)
	if err != nil {
		return err
	}
	return nil
}

func (twitch *TwitchUserClient) RefreshAccessToken() error {
	req_url := "https://id.twitch.tv/oauth2/token"

	values := url.Values{}
	values.Set("grant_type", "refresh_token")
	values.Add("refresh_token", twitch.RefreshToken)
	values.Add("client_id", twitch.ClientID)
	values.Add("client_secret", twitch.ClientSecret)

	req, err := http.NewRequest(
		"POST",
		req_url,
		strings.NewReader(values.Encode()),
	)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := new(http.Client)
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("Error: %s", resp.Status)
	}

	byteArray, _ := ioutil.ReadAll(resp.Body)
	var mapBody map[string]interface{}
	json.Unmarshal(byteArray, &mapBody)
	access_token := mapBody["access_token"].(string)
	refresh_token := mapBody["refresh_token"].(string)
	if err := twitch.UpdateSocialToken(access_token, refresh_token); err != nil {
		return err
	}
	return nil
}

func (twitch *TwitchUserClient) UpdateSocialToken(access_token, refresh_token string) error {
	twitch.AccessToken = access_token
	twitch.RefreshToken = refresh_token

	user, err := user.GetByUsername(twitch.Username)
	if err != nil {
		return err
	}
	socialaccount, err := socialaccount.GetByUserId(user.ID)
	if err != nil {
		return err
	}
	socialtoken, err := socialtoken.GetBySocialaccountId(socialaccount.ID)
	if err != nil {
		return err
	}
	socialtoken.AccessToken = access_token
	socialtoken.RefreshToken = refresh_token
	db.GetDB().Save(&socialtoken)
	return nil
}

func (tc *TwitchUserClient) Request(method string, url string, data *map[string]interface{}) (map[string]interface{}, string, int) {
	b := new(bytes.Buffer)
	json.NewEncoder(b).Encode(data)
	req, _ := http.NewRequest(
		method,
		url,
		nil,
	)
	req.Header.Add("Authorization", "Bearer "+tc.AccessToken)
	req.Header.Add("Client-ID", tc.ClientID)

	client := new(http.Client)
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 && resp.StatusCode != 202 {
		tc.Count++
		if tc.Count <= 1 {
			if resp.StatusCode == 401 {
				err := tc.RefreshAccessToken()
				if err != nil {
					log.Println(err)
					return nil, "", resp.StatusCode
				}
				log.Printf("User[%s]'s token refreshed\n", tc.Username)
				return tc.Request(method, url, data)
			} else {
				return tc.Request(method, url, data)
			}
		} else {
			log.Println(resp.Status, req.Header, req.URL, req.Method, req.Body)
			return nil, "", resp.StatusCode
		}
	} else {
		byteArray, _ := ioutil.ReadAll(resp.Body)
		headers := resp.Header
		ratelimit_remaining := headers["Ratelimit-Remaining"][0]
		log.Println(headers)
		var mapBody map[string]interface{}
		json.Unmarshal(byteArray, &mapBody)
		return mapBody["data"].([]interface{})[0].(map[string]interface{}), ratelimit_remaining, resp.StatusCode
	}
}

func AuthConfig() *oauth2.Config {
	conf := &oauth2.Config{
		ClientID:     os.Getenv("TWITCH_CLIENT_ID"),
		ClientSecret: os.Getenv("TWITCH_CLIENT_SECRET"),
		Scopes:       []string{"user:read:email", "chat:read", "moderator:read:chat_settings", "clips:edit"},
		RedirectURL:  os.Getenv("APP_BASE_URL") + "/account/twitch/login/callback/",
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://id.twitch.tv/oauth2/authorize",
			TokenURL: "https://id.twitch.tv/oauth2/token",
		},
	}
	return conf
}

func State(n int) (string, error) {
	data := make([]byte, n)
	if _, err := io.ReadFull(rand.Reader, data); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(data), nil
}

func RedirectURL() (string, string, error) {
	conf := AuthConfig()

	// Redirect user to consent page to ask for permission
	// for the scopes specified above.
	state, err := State(10)
	if err != nil {
		return "", "", err
	}
	url := conf.AuthCodeURL(state, oauth2.AccessTypeOffline)
	return url, state, nil
}

func AccessToken(code string) (*oauth2.Token, error) {
	conf := AuthConfig()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Use the authorization code that is pushed to the redirect URL.
	// Exchange will do the handshake to retrieve the initial access token.
	// The HTTP Client returned by conf.
	// Client will refresh the token as necessary.
	tok, err := conf.Exchange(ctx, code)
	if err != nil {
		log.Println(code, tok, conf)
		log.Fatal(err)
		return nil, err
	}
	return tok, nil
}

func UpdateTokenInfo(info map[string]interface{}, token *oauth2.Token) error {
	username := info["login"].(string)
	email := info["email"].(string)

	u, err := user.GetByUsername(username)
	if err != nil {
		u, err = user.Create(username, email)
		if err != nil {
			return err
		}
	}

	name := "Twitch"
	p, err := provider.GetByName(name)
	if err != nil {
		p, err = provider.Create(name)
		if err != nil {
			return err
		}
	}

	sa, err := socialaccount.GetByUserId(u.ID)
	if err != nil {
		sa, err = socialaccount.Create(u, p, info)
		if err != nil {
			return err
		}
	}

	st, err := socialtoken.GetBySocialaccountId(sa.ID)
	if err != nil {
		st, err = socialtoken.Create(sa, p, token)
		if err != nil {
			return err
		}
	}
	st.AccessToken = token.AccessToken
	st.RefreshToken = token.RefreshToken
	st.Expiry = token.Expiry
	db.GetDB().Save(&st)

	return nil
}
