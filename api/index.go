package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/nikoksr/notify"
	"github.com/nikoksr/notify/service/bark"
	"github.com/nikoksr/notify/service/discord"
	"github.com/nikoksr/notify/service/telegram"
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
	if event.Action != "created" {
		w.WriteHeader(http.StatusOK)
		return
	}
	title := fmt.Sprintf("New GitHub Star on %s", escape(event.Repository.FullName))
	text := fmt.Sprintf(
		"[%s](%s) starred [%s](%s), now it has **%d** stars\\.",
		escape(event.Sender.Login),
		escape(event.Sender.HtmlUrl),
		escape(event.Repository.FullName),
		escape(event.Repository.HtmlUrl),
		event.Repository.StarGazersCount,
	)

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()
	err = Notify(ctx, title, text)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("OK"))
}

func envOrDefault(key, defaultValue string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return defaultValue
}

func env(key string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	panic(fmt.Sprintf("environment variable %s is not set", key))
}

func Notify(ctx context.Context, subject, message string) (err error) {
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("notify: %v", e)
		}
	}()

	notifier := notify.New()
	service := envOrDefault("NOTIFY_SERVICE", "bark")
	switch service {
	case "bark":
		notifier.UseServices(bark.NewWithServers(env("BARK_KEY"), envOrDefault("BARK_SERVER", bark.DefaultServerURL)))
	case "telegram":
		tg, err := telegram.New(env("TELEGRAM_TOKEN"))
		if err != nil {
			return fmt.Errorf("telegram: %w", err)
		}
		tg.SetParseMode("MarkdownV2")
		chatIDStr := env("TELEGRAM_CHAT_ID")
		chatID, err := strconv.ParseInt(chatIDStr, 10, 64)
		if err != nil {
			return fmt.Errorf("telegram parse chat_id: %w", err)
		}
		tg.AddReceivers(chatID)
		notifier.UseServices(tg)
	case "discord":
		ds := discord.New()
		ds.AddReceivers(env("DISCORD_CHANNEL_ID"))
		notifier.UseServices(ds)
	}
	err = notifier.Send(ctx, subject, message)
	if err != nil {
		return fmt.Errorf("send: %w", err)
	}
	return nil
}

var replacer = strings.NewReplacer(
	"_", "\\_", "*", "\\*", "[", "\\[", "]", "\\]", "(",
	"\\(", ")", "\\)", "~", "\\~", "`", "\\`", ">", "\\>",
	"#", "\\#", "+", "\\+", "-", "\\-", "=", "\\=", "|",
	"\\|", "{", "\\{", "}", "\\}", ".", "\\.", "!", "\\!",
	"\\", "\\\\",
)

func escape(text string) string {
	return replacer.Replace(text)
}
