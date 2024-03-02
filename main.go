package main

import (
	"YT-Spotify-Favourite-Sync/spotify"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strconv"
	"strings"
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
	var spc *spotify.Client

	state := rand.Int()
	http.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		sendBody := fmt.Sprintf("response_type=code&client_id=%v&scope=user-library-read,user-library-modify&redirect_uri=http://localhost:8866/spotify&state=%v", c.Spc.Id, state)
		sendBody = url.PathEscape(sendBody)
		http.Redirect(writer, request, "https://accounts.spotify.com/authorize?"+sendBody, http.StatusSeeOther)
	})

	http.HandleFunc("/spotify", func(writer http.ResponseWriter, request *http.Request) {
		x := strings.Split(strings.SplitAfterN(request.RequestURI, "=", 2)[1], "&")
		authToken := x[0]
		state2 := strings.Split(x[1], "=")[1]
		if strconv.Itoa(state) != state2 {
			log.Fatalf("SPOTIFY: Request has been tampered with by a third party\nstate expected: %v\nfound: %v\naborting!", state, state2)
		}

		spc = spotify.NewClient(c.Spc, authToken)
	})

	go func() {
		_ = exec.Command("xdg-open", "http://localhost:8866/").Start()
		err = http.ListenAndServe(":8866", nil)
		if err != nil {
			return
		}
	}()

	for spc == nil {
	}

	tracks := spc.FindSavedTracks()
	for _, track := range tracks {
		fmt.Println(track)
	}
}
