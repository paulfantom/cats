package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type Cat struct {
	Id         string            `json:"id,omitempty"`
	Url        string            `json:"url"`
	Width      int               `json:"width,omitempty"`
	Height     int               `json:"height,omitempty"`
	Breeds     []string          `json:"breeds,omitempty"`
	Faviourite map[string]string `json:"favourite,omitempty"`
}

func getCat(apiKey string) (string, error) {
	url := "https://api.thecatapi.com/v1/images/search"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Add("x-api-key", apiKey)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return "", err
	}

	var cats []Cat
	json.Unmarshal(body, &cats)

	return cats[0].Url, nil
}

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	http.HandleFunc("/random", func(w http.ResponseWriter, r *http.Request) {
		log.Info().Msg("Request received")
		apiKey := r.Header.Get("x-api-key")
		cat, err := getCat(apiKey)
		if err != nil {
			log.Error().Err(err).Msg("Error getting a cat from the external API")
			http.Error(w, "Error", http.StatusInternalServerError)
		}
		log.Info().Msg("Sent a cat to client")
		http.Redirect(w, r, cat, http.StatusFound)
	})

	log.Info().Msg("Listening on port 8080")
	http.ListenAndServe(":8080", nil)
}
