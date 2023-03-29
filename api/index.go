package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/j178/github_stargazer/notify"
	"github.com/j178/github_stargazer/utils"
)

type StarEvent struct {
	Action     string `json:"action"`
	Repository struct {
		Name            string `json:"name"`
		FullName        string `json:"full_name"`
		HtmlUrl         string `json:"html_url"`
		StarGazersCount int    `json:"stargazers_count"`
	} `json:"repository"`
	Sender struct {
		Login   string `json:"login"`
		HtmlUrl string `json:"html_url"`
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
	var title, text string
	switch event.Action {
	case "deleted":
		title = fmt.Sprintf("Lost GitHub Star on %s", utils.EscapeMarkdown(event.Repository.FullName))
		text = fmt.Sprintf(
			"[%s](%s) unstarred [%s](%s), now it has **%d** stars\\.",
			utils.EscapeMarkdown(event.Sender.Login),
			utils.EscapeMarkdown(event.Sender.HtmlUrl),
			utils.EscapeMarkdown(event.Repository.FullName),
			utils.EscapeMarkdown(event.Repository.HtmlUrl),
			event.Repository.StarGazersCount,
		)
	case "created":
		title = fmt.Sprintf("New GitHub Star on %s", utils.EscapeMarkdown(event.Repository.FullName))
		text = fmt.Sprintf(
			"[%s](%s) starred [%s](%s), now it has **%d** stars\\.",
			utils.EscapeMarkdown(event.Sender.Login),
			utils.EscapeMarkdown(event.Sender.HtmlUrl),
			utils.EscapeMarkdown(event.Repository.FullName),
			utils.EscapeMarkdown(event.Repository.HtmlUrl),
			event.Repository.StarGazersCount,
		)
	}

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()
	err = notify.Notify(ctx, title, text)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("OK"))
}
