package spotify

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

const BaseUrl = "https://api.spotify.com/v1/"

type Config struct {
	Id           string
	Secret       string
	PlayListName string
}

type Client struct {
	cfg   *Config
	token struct {
		sync.Mutex
		value string
	}
}

type tokenResponse struct {
	Access_token  string
	Scope         string
	Expires_in    int
	Refresh_token string
}

func NewClient(cfg Config, auth_token string) *Client {
	//get access Token
	r, err := http.NewRequest("POST", "https://accounts.spotify.com/api/token", strings.NewReader(fmt.Sprintf("grant_type=authorization_code&code=%v&redirect_uri=http://localhost:8866/spotify", auth_token)))
	if err != nil {
		log.Fatalf("SPOTIFY: failed to build access token request\nerr: %v", err)
	}

	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	r.Header.Add("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%v:%v", cfg.Id, cfg.Secret))))

	response, err := http.DefaultClient.Do(r)
	if err != nil {
		log.Fatalf("SPOTIFY: failed to send request to get access token\nerr:%v", err)
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		log.Fatalf("SPOTIFY: failed to parse access token request body\nerr:%v\nstatus:%v", err, response.Status)
	}

	if response.StatusCode != 200 {
		log.Fatalf("SPOTIFY: failed to get access token\nbody:%v\nstatus:%v", string(body), response.Status)
	}

	var parsed tokenResponse
	err = json.Unmarshal(body, &parsed)
	if err != nil {
		log.Fatalf("SPOTIFY: failed to parse access token response body\nbody:%v\nerror:%v", body, err)
	}

	c := Client{
		cfg: &cfg,
		token: struct {
			sync.Mutex
			value string
		}{value: parsed.Access_token},
	}

	//refresh tokens
	go c.refreshAccess(parsed.Expires_in, parsed.Refresh_token)

	return &c
}

func (c *Client) refreshAccess(expiresIn int, refreshToken string) {
	duration, err := time.ParseDuration(fmt.Sprintf("%vs", expiresIn-60))
	if err != nil {
		log.Printf("SPOTIFY: failed to parse expires in duration")
		return
	}
	time.Sleep(duration)

	r, err := http.NewRequest("POST", BaseUrl, strings.NewReader(fmt.Sprintf("grant_type=refresh_token&refresh_token=%v", refreshToken)))
	if err != nil {
		log.Fatalf("SPOTIFY: failed to build refresh token request\nerr: %v", err)
	}

	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	r.Header.Add("Authorization", base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%v,%v", c.cfg.Id, c.cfg.Secret))))

	response, err := http.DefaultClient.Do(r)
	if err != nil {
		log.Fatalf("SPOTIFY: failed to send request to refresh token\nerr:%v", err)
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		log.Fatalf("SPOTIFY: failed to parse refresh token request body\nerr:%v\nstatus:%v", err, response.Status)
	}

	if response.StatusCode != 200 {
		log.Fatalf("SPOTIFY: failed to get token refresh \nbody:%v\nstatus:%v", string(body), response.Status)
	}

	var parsed tokenResponse
	err = json.Unmarshal(body, &parsed)
	if err != nil {
		log.Fatalf("SPOTIFY: failed to parse refresh token response body\nbody:%v\nerror:%v", body, err)
	}

	c.token.Lock()
	c.token.value = parsed.Access_token
	c.token.Unlock()

	c.refreshAccess(parsed.Expires_in, parsed.Refresh_token)
}

type myTracksResp struct {
	Next  string
	Items []myTracksRespItem
}

type myTracksRespItem struct {
	AddedAt string
	Track   struct {
		Name string
	}
}

func (c *Client) doFind(url string, out *[]myTracksRespItem) {
	// find playlist
	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Fatal("SPOTIFY:" + err.Error())
	}

	c.token.Lock()
	request.Header.Add("Authorization", "Bearer "+c.token.value)

	response, err := http.DefaultClient.Do(request)
	c.token.Unlock()
	if err != nil {
		log.Fatal("SPOTIFY:" + err.Error())
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		log.Fatal("SPOTIFY:" + err.Error())
	}

	if response.StatusCode != 200 {
		log.Fatalf("SPOTIFY: failed to get playlists\nStatus:" + response.Status + "\nbody:" + string(body))
	}

	var m myTracksResp
	err = json.Unmarshal(body, &m)
	if err != nil {
		log.Fatal("SPOTIFY:" + err.Error())
	}

	*out = append(*out, m.Items...)

	if m.Next != "" {
		c.doFind(m.Next, out)
	}
}

func (c *Client) FindSavedTracks() []myTracksRespItem {
	var tracks []myTracksRespItem
	c.doFind(BaseUrl+"me/tracks?limit=50&offset=0", &tracks)
	return tracks
}
