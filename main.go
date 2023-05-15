package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
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
	http.HandleFunc("/random", func(w http.ResponseWriter, r *http.Request) {
		apiKey := r.Header.Get("x-api-key")
		cat, err := getCat(apiKey)
		if err != nil {
			http.Error(w, "Error", http.StatusInternalServerError)
		}
		http.Redirect(w, r, cat, http.StatusFound)
	})

	http.ListenAndServe(":8080", nil)
}
