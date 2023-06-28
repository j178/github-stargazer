package api

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/redis/rueidis"

	"github.com/j178/github_stargazer/config"
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

func validateSignature(r *http.Request) bool {
	if config.WebhookSecret == "" {
		return true
	}

	payload, err := io.ReadAll(r.Body)
	if err != nil {
		return false
	}

	// Restore the io.ReadCloser to its original state so it can be read later
	r.Body = io.NopCloser(bytes.NewBuffer(payload))

	signature := r.Header.Get("X-Hub-Signature-256")
	if signature == "" {
		return false
	}

	mac := hmac.New(sha256.New, []byte(config.WebhookSecret))
	mac.Write(payload)
	expectedMAC := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(signature[len("sha256="):]), []byte(expectedMAC))
}

type Setting struct {
	NotifySettings []map[string]string `json:"notify_settings"`
	AllowRepos     []string            `json:"allow_repos"`
	MuteRepos      []string            `json:"mute_repos"`
}

func (s *Setting) IsAllowRepo(fullName string) bool {
	for _, repo := range s.MuteRepos {
		if repo == fullName {
			return false
		}
	}
	if len(s.AllowRepos) == 0 {
		return true
	}
	for _, repo := range s.AllowRepos {
		if repo == fullName {
			return true
		}
	}
	return false
}

var redis rueidis.Client

func init() {
	u, err := url.Parse(config.KvURL)
	if err != nil {
		panic(err)
	}
	passwd, _ := u.User.Password()
	opt := rueidis.ClientOption{
		Username: u.User.Username(),
		Password: passwd,
		InitAddress: []string{
			u.Host,
		},
	}
	if u.Scheme == "rediss" {
		opt.TLSConfig = &tls.Config{
			ServerName: u.Host,
		}
	}
	redis, err = rueidis.NewClient(opt)
	if err != nil {
		panic(err)
	}
}

func getSettings(login string) (*Setting, error) {
	ctx := context.Background()
	s, err := redis.Do(ctx, redis.B().Get().Key("github_stargazer:settings:"+login).Build()).AsBytes()
	if err != nil {
		return nil, err
	}
	var setting Setting
	err = json.Unmarshal(s, &setting)
	if err != nil {
		return nil, err
	}
	return &setting, nil
}

func saveSettings(login string, setting Setting) error {
	b, err := json.Marshal(setting)
	if err != nil {
		return err
	}

	err = redis.Do(
		context.Background(),
		redis.B().Set().Key("github_stargazer:settings:"+login).Value(string(b)).Build(),
	).Error()
	if err != nil {
		return err
	}
	return nil
}

func compose(evt *StarEvent) (string, string) {
	var title, text string
	switch evt.Action {
	case "deleted":
		title = fmt.Sprintf("Lost GitHub Star on %s", utils.EscapeMarkdown(evt.Repository.FullName))
		text = fmt.Sprintf(
			"[%s](%s) unstarred [%s](%s), now it has **%d** stars\\.",
			utils.EscapeMarkdown(evt.Sender.Login),
			utils.EscapeMarkdown(evt.Sender.HtmlUrl),
			utils.EscapeMarkdown(evt.Repository.FullName),
			utils.EscapeMarkdown(evt.Repository.HtmlUrl),
			evt.Repository.StarGazersCount,
		)
	case "created":
		title = fmt.Sprintf("New GitHub Star on %s", utils.EscapeMarkdown(evt.Repository.FullName))
		text = fmt.Sprintf(
			"[%s](%s) starred [%s](%s), now it has **%d** stars\\.",
			utils.EscapeMarkdown(evt.Sender.Login),
			utils.EscapeMarkdown(evt.Sender.HtmlUrl),
			utils.EscapeMarkdown(evt.Repository.FullName),
			utils.EscapeMarkdown(evt.Repository.HtmlUrl),
			evt.Repository.StarGazersCount,
		)
	}
	return title, text
}

func OnEvent(w http.ResponseWriter, r *http.Request) {
	if !validateSignature(r) {
		http.Error(w, "Bad signature", http.StatusForbidden)
		return
	}

	if r.Header.Get("X-GitHub-Event") != "star" {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("Not a star event"))
		return
	}

	var event StarEvent
	err := json.NewDecoder(r.Body).Decode(&event)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	settings, err := getSettings(event.Sender.Login)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	if !settings.IsAllowRepo(event.Repository.FullName) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("Repo not allowed"))
		return
	}

	notifier, err := notify.GetNotifier(settings.NotifySettings)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	title, content := compose(&event)

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()
	err = notifier.Send(ctx, title, content)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("Sent"))
}
