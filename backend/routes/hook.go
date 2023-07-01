package routes

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/go-github/v53/github"
	"github.com/sourcegraph/conc/pool"

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
		Abort(c, http.StatusBadRequest, err, "validate payload")
		return
	}

	webhookType := github.WebHookType(c.Request)
	event, err := github.ParseWebHook(webhookType, payload)
	if err != nil {
		Abort(c, http.StatusBadRequest, err, "parse webhook")
		return
	}

	switch webhookType {
	case "star":
		evt, _ := event.(*github.StarEvent)
		settings, err := cache.GetAllSettings(c, evt.Repo.Owner.GetLogin())
		if err != nil {
			Abort(c, http.StatusInternalServerError, err, "get all settings")
			return
		}
		if len(settings) == 0 {
			c.String(http.StatusOK, "Settings not found")
			return
		}

		wg := pool.New().WithContext(c).WithMaxGoroutines(10)
		for _, setting := range settings {
			setting := setting
			if !setting.IsAllowRepo(evt.Repo.GetFullName()) {
				continue
			}
			wg.Go(
				func(ctx context.Context) error {
					return sendNotify(ctx, evt, setting)
				},
			)
		}
		err = wg.Wait()
		if err != nil {
			Abort(c, http.StatusInternalServerError, err, "notify")
			return
		}

	default:
		c.String(http.StatusOK, "Not interested event")
	}
	return
}

func sendNotify(ctx context.Context, evt *github.StarEvent, setting *cache.Setting) error {
	notifier, err := notify.GetNotifier(setting.NotifySettings)
	if err != nil {
		return err
	}

	title, content := compose(evt)

	err = notifier.Send(ctx, title, content)
	if err != nil {
		return err
	}
	return nil
}
