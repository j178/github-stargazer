package configure

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/go-github/v84/github"
	"github.com/samber/lo"

	"github.com/j178/github_stargazer/backend/cache"
	"github.com/j178/github_stargazer/backend/notify"
	"github.com/j178/github_stargazer/backend/routes"
	"github.com/j178/github_stargazer/backend/utils"
)

const MaxSettingsCount = 10

const (
	installedReposPerPage = 30
	repoSearchLimit       = 20
)

type repoPayload struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Fork        bool   `json:"fork"`
}

func GetSettings(c *gin.Context) {
	login := c.GetString("login")
	account := c.Param("account")

	settings, err := cache.GetSettings(c, account, login)
	if err != nil {
		routes.Abort(c, http.StatusNotFound, err, "")
		return
	}

	c.JSON(http.StatusOK, settings)
}

func UpdateSettings(c *gin.Context) {
	login := c.GetString("login")
	account := c.Param("account")

	if !checkAccountAssociation(c, account, login) {
		return
	}

	var setting cache.Setting
	err := c.ShouldBindJSON(&setting)
	if err != nil {
		routes.Abort(c, http.StatusBadRequest, err, "")
		return
	}

	if len(setting.NotifySettings) > MaxSettingsCount {
		routes.Abort(c, http.StatusBadRequest, nil, fmt.Sprintf("max settings count is %d", MaxSettingsCount))
		return
	}

	// check notify settings (check token is valid too)
	_, err = notify.GetNotifier(setting.NotifySettings)
	if err != nil {
		routes.Abort(c, http.StatusBadRequest, err, "invalid notify settings")
		return
	}

	err = cache.SaveSettings(c, account, login, setting)
	if err != nil {
		routes.Abort(c, http.StatusInternalServerError, err, "save settings")
		return
	}

	c.JSON(http.StatusOK, gin.H{})
}

func DeleteSettings(c *gin.Context) {
	login := c.GetString("login")
	account := c.Param("account")

	err := cache.DeleteSettings(c, account, login)
	if err != nil {
		routes.Abort(c, http.StatusInternalServerError, err, "delete settings")
		return
	}

	c.JSON(http.StatusOK, gin.H{})
}

func getInstallations(ctx context.Context, login string) ([]*github.Installation, error) {
	token, err := cache.GetOAuthToken(ctx, login)
	if err != nil {
		return nil, err
	}

	client := github.NewTokenClient(ctx, token)
	opts := &github.ListOptions{PerPage: 100}

	var installations []*github.Installation
	for {
		ins, resp, err := client.Apps.ListUserInstallations(ctx, opts)
		if err != nil {
			return nil, err
		}
		installations = append(installations, ins...)
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return installations, nil
}

func getInstallationAccounts(ctx context.Context, login string) ([]string, error) {
	l, err := getInstallations(ctx, login)
	if err != nil {
		return nil, err
	}
	accounts := lo.Map(
		l, func(item *github.Installation, _ int) string {
			return item.Account.GetLogin()
		},
	)
	return accounts, nil
}

// check account is associated with login
func checkAccountAssociation(c *gin.Context, account, login string) bool {
	installations, err := cache.GetOrCreate(
		c, cache.Key{"installations", login}, 24*time.Hour, func() ([]string, error) {
			return getInstallationAccounts(c, login)
		},
	)
	if err != nil {
		routes.Abort(c, http.StatusInternalServerError, err, "get installations")
		return false
	}
	if !lo.Contains(installations, account) {
		routes.Abort(
			c,
			http.StatusForbidden,
			nil,
			fmt.Sprintf("app not installed to %s, or you have no permission", account),
		)
		return false
	}
	return true
}

func Installations(c *gin.Context) {
	login := c.GetString("login")

	installations, err := getInstallations(c, login)
	if err != nil {
		routes.Abort(c, http.StatusInternalServerError, err, "get installations")
		return
	}

	result := make([]map[string]any, len(installations))
	accounts := make([]string, len(installations))
	for i, item := range installations {
		result[i] = map[string]any{
			"id":           item.GetID(),
			"account":      item.Account.GetLogin(),
			"account_type": item.Account.GetType(),
		}
		accounts[i] = item.Account.GetLogin()
	}

	_ = cache.Set(c, cache.Key{"installations", login}, accounts, 24*time.Hour)

	c.JSON(http.StatusOK, result)
}

func parseInstallationID(c *gin.Context) (int64, bool) {
	installationIDStr := c.Param("installationID")
	installationID, err := strconv.ParseInt(installationIDStr, 10, 64)
	if err != nil {
		routes.Abort(c, http.StatusBadRequest, nil, "invalid installationID")
		return 0, false
	}
	return installationID, true
}

func listInstallationRepos(ctx context.Context, installationID int64) ([]repoPayload, error) {
	return cache.GetOrCreate(
		ctx,
		cache.Key{"installation_repos", strconv.FormatInt(installationID, 10)},
		10*time.Minute,
		func() ([]repoPayload, error) {
			token, err := cache.GetInstallationToken(ctx, installationID)
			if err != nil {
				return nil, err
			}

			client := github.NewTokenClient(ctx, token)
			opts := &github.ListOptions{PerPage: 100}
			repos := make([]repoPayload, 0, 10)

			for {
				list, resp, err := client.Apps.ListRepos(ctx, opts)
				if err != nil {
					return nil, err
				}
				repos = append(repos, serializeRepos(list.Repositories)...)
				if resp.NextPage == 0 {
					break
				}
				opts.Page = resp.NextPage
			}

			return repos, nil
		},
	)
}

func serializeRepos(repos []*github.Repository) []repoPayload {
	returnRepos := make([]repoPayload, len(repos))
	for i, item := range repos {
		returnRepos[i] = repoPayload{
			ID:          item.GetID(),
			Name:        item.GetFullName(),
			Description: item.GetDescription(),
			Fork:        item.GetFork(),
		}
	}
	return returnRepos
}

func InstalledRepos(c *gin.Context) {
	installationID, ok := parseInstallationID(c)
	if !ok {
		return
	}
	page := c.DefaultQuery("page", "1")
	pageInt, err := strconv.Atoi(page)
	if err != nil {
		routes.Abort(c, http.StatusBadRequest, nil, "invalid page")
		return
	}
	if pageInt < 1 {
		routes.Abort(c, http.StatusBadRequest, nil, "invalid page")
		return
	}

	repos, err := listInstallationRepos(c, installationID)
	if err != nil {
		routes.Abort(c, http.StatusInternalServerError, err, "list repos")
		return
	}

	start := (pageInt - 1) * installedReposPerPage
	if start >= len(repos) {
		c.JSON(http.StatusOK, []repoPayload{})
		return
	}
	end := min(start+installedReposPerPage, len(repos))
	c.JSON(http.StatusOK, repos[start:end])
}

func SearchInstalledRepos(c *gin.Context) {
	installationID, ok := parseInstallationID(c)
	if !ok {
		return
	}

	query := strings.TrimSpace(c.Query("q"))
	if query == "" {
		c.JSON(http.StatusOK, []repoPayload{})
		return
	}

	limit := repoSearchLimit
	if limitStr := strings.TrimSpace(c.Query("limit")); limitStr != "" {
		value, err := strconv.Atoi(limitStr)
		if err != nil || value < 1 || value > 100 {
			routes.Abort(c, http.StatusBadRequest, nil, "invalid limit")
			return
		}
		limit = value
	}

	repos, err := listInstallationRepos(c, installationID)
	if err != nil {
		routes.Abort(c, http.StatusInternalServerError, err, "search repos")
		return
	}

	normalizedQuery := strings.ToLower(query)
	matches := make([]repoPayload, 0, limit)
	for _, repo := range repos {
		if !strings.Contains(strings.ToLower(repo.Name), normalizedQuery) {
			continue
		}
		matches = append(matches, repo)
		if len(matches) >= limit {
			break
		}
	}

	c.JSON(http.StatusOK, matches)
}

func CheckSettings(c *gin.Context) {
	var setting cache.Setting
	err := c.ShouldBindJSON(&setting)
	if err != nil {
		routes.Abort(c, http.StatusBadRequest, err, "")
		return
	}

	if len(setting.NotifySettings) > MaxSettingsCount {
		routes.Abort(c, http.StatusBadRequest, nil, fmt.Sprintf("max settings count is %d", MaxSettingsCount))
		return
	}

	_, err = notify.GetNotifier(setting.NotifySettings)
	if err != nil {
		routes.Abort(c, http.StatusBadRequest, err, "invalid notify settings")
		return
	}

	c.JSON(http.StatusOK, gin.H{})
}

func TestNotify(c *gin.Context) {
	var setting cache.Setting
	err := c.ShouldBindJSON(&setting)
	if err != nil {
		routes.Abort(c, http.StatusBadRequest, err, "")
		return
	}

	if len(setting.NotifySettings) > MaxSettingsCount {
		routes.Abort(c, http.StatusBadRequest, nil, fmt.Sprintf("max settings count is %d", MaxSettingsCount))
		return
	}

	notifier, err := notify.GetNotifier(setting.NotifySettings)
	if err != nil {
		routes.Abort(c, http.StatusBadRequest, err, "invalid notify settings")
		return
	}

	err = notifier.Send(c, "Test Message", utils.EscapeMarkdown("This is a test message from https://github-stargazer.vercel.app/"))
	if err != nil {
		routes.Abort(c, http.StatusInternalServerError, err, "send test notify")
		return
	}

	c.JSON(http.StatusOK, gin.H{})
}
