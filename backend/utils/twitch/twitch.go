package twitch

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/halmk/cliplist-ttv/backend/db"
	"github.com/halmk/cliplist-ttv/backend/entity"
	"github.com/jinzhu/gorm/dialects/postgres"
	"golang.org/x/oauth2"
)

type TwitchAppClient struct {
	client_id     string
	client_secret string
	token         string
	count         int
}

type TwitchUserClient struct {
	client_id string
	token     string
	count     int
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

func NewTwitchUserClient(token string) TwitchUserClient {
	client_id := os.Getenv("TWITCH_CLIENT_ID")
	twitch := TwitchUserClient{
		client_id,
		token,
		0,
	}
	return twitch
}

func (twitch *TwitchAppClient) GetToken() {
	url := "https://id.twitch.tv/oauth2/token"
	url += "?client_id=" + twitch.client_id + "&client_secret=" + twitch.client_secret + "&grant_type=client_credentials"
	req, _ := http.NewRequest(
		"POST",
		url,
		nil,
	)
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
		if twitch.count < 2 {
			twitch.GetToken()
			return twitch.GetRequest(url)
		} else {
			return make(map[string]interface{}), resp.StatusCode
		}
	} else {
		byteArray, _ := ioutil.ReadAll(resp.Body)
		var mapBody map[string]interface{}
		json.Unmarshal(byteArray, &mapBody)
		return mapBody, resp.StatusCode
	}
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

func (tc *TwitchUserClient) GetRequest(url string) (map[string]interface{}, int) {
	req, _ := http.NewRequest(
		"GET",
		url,
		nil,
	)
	req.Header.Add("Authorization", "Bearer "+tc.token)
	req.Header.Add("Client-ID", tc.client_id)

	client := new(http.Client)
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Println(resp.Status, req.Header, req.URL)
		return make(map[string]interface{}), resp.StatusCode
		}
	} else {
		byteArray, _ := ioutil.ReadAll(resp.Body)
		var mapBody map[string]interface{}
		json.Unmarshal(byteArray, &mapBody)
		return mapBody["data"].([]interface{})[0].(map[string]interface{}), resp.StatusCode
	}
}

func AuthConfig() *oauth2.Config {
	conf := &oauth2.Config{
		ClientID:     os.Getenv("TWITCH_CLIENT_ID"),
		ClientSecret: os.Getenv("TWITCH_CLIENT_SECRET"),
		Scopes:       []string{"user:read:email"},
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
