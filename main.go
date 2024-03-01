package main

import (
	"YT-Spotify-Favourite-Sync/spotify"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
)

type config struct {
	Spc spotify.Config
}

func loadConfig() (config, error) {
	cfg := config{}

	file, err := os.ReadFile("config.json")
	if err != nil {
		return config{}, err
	}

	err = json.Unmarshal(file, &cfg)
	if err != nil {
		return config{}, err
	}

	return cfg, nil
}

func main() {
	c, err := loadConfig()
	if err != nil {
		log.Fatal("Failed to load config:\n" + err.Error())
	}

	http.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		state := rand.Int()
		sendBody := fmt.Sprintf("response_type=code&client_id=%v&scope=playlist-modify-private,playlist-read-private&redirect_uri=http://localhost:8866/spotify&state=%v", c.Spc.Id, state)
		sendBody = url.PathEscape(sendBody)
		http.Redirect(writer, request, "https://accounts.spotify.com/authorize?"+sendBody, http.StatusSeeOther)
	})

	http.HandleFunc("/spotify", func(writer http.ResponseWriter, request *http.Request) {
		fmt.Println(io.ReadAll(request.Body))
		//TODO
	})

	err = http.ListenAndServe(":8866", nil)
	if err != nil {
		return
	}

	spc := spotify.NewClient(c.Spc, "rt")
	for _, track := range spc.FindPlayListTracks() {
		fmt.Println(track)
	}
}
