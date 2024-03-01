package spotify

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
)

const baseUrl = "https://api.spotify.com/v1/"

type Config struct {
	Id           string
	PlayListName string
}

type Client struct {
	cfg   *Config
	token string
}

func NewClient(cfg Config, token string) *Client {
	c := Client{
		cfg:   &cfg,
		token: token,
	}

	return &c
}

func (c *Client) FindPlayListTracks() []string {

	// find playlist
	request, err := http.NewRequest(http.MethodGet, baseUrl+"me/playlists?limit=50&offset=0", nil)
	if err != nil {
		log.Fatal("SPOTIFY:" + err.Error())
	}

	request.Header.Add("Authorization", "Bearer "+c.token)

	response, err := http.DefaultClient.Do(request)
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

	m := struct {
		Items []struct {
			Name   string
			Id     string
			Tracks struct {
				Href  string
				Total int
			}
		}
	}{}

	err = json.Unmarshal(body, &m)
	if err != nil {
		log.Fatal("SPOTIFY:" + err.Error())
	}

	// find tracks
	playListHref := ""
	for _, item := range m.Items {
		if item.Name == c.cfg.PlayListName {
			playListHref = item.Tracks.Href
			break
		}
	}

	if playListHref == "" {
		log.Fatal("SPOTIFY: failed to get href to tracks\n body:" + string(body))
	}

	var tracks []string

	for playListHref != "" {
		request, err = http.NewRequest(http.MethodGet, playListHref, nil)
		if err != nil {
			log.Fatal("SPOTIFY:" + err.Error())
		}
		request.Header.Add("Authorization", "Bearer "+c.token)

		n := struct {
			Next  string
			Items []struct {
				Name string
				Id   string
			}
		}{}

		err = json.Unmarshal(body, &n)
		if err != nil {
			log.Fatal("SPOTIFY:" + err.Error())
		}

		for _, i := range n.Items {
			tracks = append(tracks, i.Name)
		}

		playListHref = n.Next
	}

	return tracks
}
