package twitch_api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

type TwitchAPI struct {
	client_id     string
	client_secret string
	token         string
	count         int
}

func NewTwitchAPI() TwitchAPI {
	client_id := os.Getenv("TWITCH_CLIENT_ID")
	client_secret := os.Getenv("TWITCH_CLIENT_SECRET")

	twitch := TwitchAPI{
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

func (twitch *TwitchAPI) GetToken() {
	url := "https://id.twitch.tv/oauth2/token"
	url += "?client_id=" + twitch.client_id + "&client_secret=" + twitch.client_secret + "&grant_type=client_credentials"
	req, _ := http.NewRequest(
		"POST",
		url,
		nil,
	)
	client := new(http.Client)
	resp, _ := client.Do(req)
	defer resp.Body.Close()

	byteArray, _ := ioutil.ReadAll(resp.Body)
	var mapBody map[string]interface{}
	json.Unmarshal(byteArray, &mapBody)
	fmt.Println(mapBody)
	twitch.token = mapBody["access_token"].(string)
	twitch.WriteTokenFile()
}

func (twitch *TwitchAPI) GetRequest(url string, params map[string]string) string {
	twitch.count++
	url += "?"
	for k, v := range params {
		url += k + "=" + v + "&"
	}
	req, _ := http.NewRequest(
		"GET",
		url,
		nil,
	)
	req.Header.Add("Authorization", "Bearer "+twitch.token)
	req.Header.Add("Client-ID", twitch.client_id)

	client := new(http.Client)
	resp, _ := client.Do(req)
	defer resp.Body.Close()

	if twitch.count < 2 && resp.StatusCode != 200 {
		twitch.GetToken()
		return twitch.GetRequest(url, params)
	} else {
		byteArray, _ := ioutil.ReadAll(resp.Body)
		var mapBody map[string]interface{}
		json.Unmarshal(byteArray, &mapBody)
		fmt.Println(mapBody)
		return string(byteArray)
	}
}

func (twitch *TwitchAPI) ReadTokenFile() error {
	b, err := ioutil.ReadFile("twitch_app_token.txt")
	if err != nil {
		twitch.GetToken()
		return twitch.WriteTokenFile()
	}
	twitch.token = string(b)
	return nil
}

func (twitch *TwitchAPI) WriteTokenFile() error {
	err := ioutil.WriteFile("twitch_app_token.txt", []byte(twitch.token), 0606)
	if err != nil {
		return err
	}
	return nil
}
