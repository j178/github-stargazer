package routes

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/go-github/v53/github"
	"github.com/redis/rueidis"

	"github.com/j178/github_stargazer/config"
	"github.com/j178/github_stargazer/notify"
	"github.com/j178/github_stargazer/utils"
)

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
var once sync.Once

func getRedis() rueidis.Client {
	once.Do(
		func() {
			u, err := url.Parse(config.KvURL)
			if err != nil {
				log.Fatalf("parse kv url failed: %s", err)
			}
			passwd, _ := u.User.Password()
			host, _, _ := net.SplitHostPort(u.Host)
			opt := rueidis.ClientOption{
				ForceSingleClient: true,
				DisableCache:      true,
				Username:          u.User.Username(),
				Password:          passwd,
				InitAddress: []string{
					u.Host,
				},
				TLSConfig: &tls.Config{
					ServerName: host,
				},
			}

			r, err := rueidis.NewClient(opt)
			if err != nil {
				log.Fatalf("create redis client failed: %s", err)
			}
			redis = r
		},
	)
	return redis
}

func getSettings(login string) (*Setting, error) {
	ctx := context.Background()
	client := getRedis()
	s, err := client.Do(ctx, client.B().Get().Key("github_stargazer:settings:"+login).Build()).AsBytes()
	if rueidis.IsRedisNil(err) {
		return nil, nil
	}
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

	client := getRedis()
	err = client.Do(
		context.Background(),
		client.B().Set().Key("github_stargazer:settings:"+login).Value(string(b)).Build(),
	).Error()
	if err != nil {
		return err
	}
	return nil
}

func compose(evt *github.StarEvent) (string, string) {
	var title, text string
	switch evt.GetAction() {
	case "deleted":
		title = fmt.Sprintf("Lost GitHub Star on %s", utils.EscapeMarkdown(evt.Repo.GetFullName()))
		text = fmt.Sprintf(
			"[%s](%s) unstarred [%s](%s), now it has **%d** stars\\.",
			utils.EscapeMarkdown(evt.Sender.GetLogin()),
			utils.EscapeMarkdown(evt.Sender.GetHTMLURL()),
			utils.EscapeMarkdown(evt.Repo.GetFullName()),
			utils.EscapeMarkdown(evt.Repo.GetHTMLURL()),
			evt.Repo.GetStargazersCount(),
		)
	case "created":
		title = fmt.Sprintf("New GitHub Star on %s", utils.EscapeMarkdown(evt.Repo.GetFullName()))
		text = fmt.Sprintf(
			"[%s](%s) starred [%s](%s), now it has **%d** stars\\.",
			utils.EscapeMarkdown(evt.Sender.GetLogin()),
			utils.EscapeMarkdown(evt.Sender.GetHTMLURL()),
			utils.EscapeMarkdown(evt.Repo.GetFullName()),
			utils.EscapeMarkdown(evt.Repo.GetHTMLURL()),
			evt.Repo.GetStargazersCount(),
		)
	}
	return title, text
}

func OnEvent(c *gin.Context) {
	payload, err := github.ValidatePayload(c.Request, []byte(config.WebhookSecret))
	if err != nil {
		c.String(http.StatusForbidden, "Bad signature")
		return
	}
	webhookType := github.WebHookType(c.Request)
	if webhookType != "star" {
		c.String(http.StatusOK, "Not a star event")
		return
	}
	event, err := github.ParseWebHook(webhookType, payload)
	if err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}
	evt, _ := event.(*github.StarEvent)

	settings, err := getSettings(evt.Sender.GetLogin())
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	if settings == nil {
		c.String(http.StatusOK, "Settings not found")
		return
	}

	if !settings.IsAllowRepo(evt.Repo.GetFullName()) {
		c.String(http.StatusOK, "Repo muted")
		return
	}

	notifier, err := notify.GetNotifier(settings.NotifySettings)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	title, content := compose(evt)

	ctx, cancel := context.WithTimeout(c, 3*time.Second)
	defer cancel()
	err = notifier.Send(ctx, title, content)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	c.String(http.StatusOK, "Sent")
}
