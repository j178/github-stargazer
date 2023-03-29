package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
)

type StarEvent struct {
	Action     string `json:"action"`
	Repository struct {
		Name            string `json:"name"`
		FullName        string `json:"full_name"`
		URL             string `json:"url"`
		StarGazersCount int    `json:"stargazers_count"`
	} `json:"repository"`
	Sender struct {
		Login string `json:"login"`
		URL   string `json:"url"`
	}
	StarredAt string `json:"starred_at"`
}

func Handler(w http.ResponseWriter, r *http.Request) {
	var event StarEvent
	err := json.NewDecoder(r.Body).Decode(&event)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if event.Action != "created" {
		w.WriteHeader(http.StatusOK)
		return
	}
	title := fmt.Sprintf("New GitHub Star on %s", event.Repository.FullName)
	text := fmt.Sprintf(
		"[%s](%s) starred [%s](%s), now it has %d stars.",
		event.Sender.Login,
		event.Sender.URL,
		event.Repository.FullName,
		event.Repository.URL,
		event.Repository.StarGazersCount,
	)
	err = bark(title, text)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("OK"))
}

var (
	barkAddress = "https://api.day.app"
)

func bark(title, text string) error {
	barkKey := os.Getenv("BARK_KEY")
	if barkKey == "" {
		return errors.New("BARK_KEY is empty")
	}
	u := barkAddress + "/" + barkKey
	if title != "" {
		title = url.QueryEscape(title)
		u = u + "/" + title
	}
	if text != "" {
		text = url.QueryEscape(text)
		u = u + "/" + text
	}
	resp, err := http.Get(u)
	if err != nil {
		return err
	}
	_ = resp.Body.Close()
	return nil
}
