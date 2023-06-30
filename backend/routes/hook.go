package routes

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/go-github/v53/github"

	"github.com/j178/github_stargazer/backend/cache"
	"github.com/j178/github_stargazer/backend/config"
	"github.com/j178/github_stargazer/backend/notify"
	"github.com/j178/github_stargazer/backend/utils"
)

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
	payload, err := github.ValidatePayload(c.Request, config.WebhookSecret)
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

	settings, err := cache.GetSettings(c, evt.Sender.GetLogin())
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
