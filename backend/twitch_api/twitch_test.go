package twitch_api

import (
	"fmt"
	"testing"
)

func TestA(t *testing.T) {
	twitch := NewTwitchAPI()
	fmt.Println(twitch)
}

func TestGetUsersA(t *testing.T) {
	twitch := NewTwitchAPI()
	url := "https://api.twitch.tv/helix/users"
	params := make(map[string]string)
	params["id"] = "141981764"
	response := twitch.GetRequest(url, params)
	fmt.Println(response)
}
