package routes

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/go-github/v53/github"
	"github.com/samber/lo"

	"github.com/j178/github_stargazer/backend/notify"

	"github.com/j178/github_stargazer/backend/cache"
)

func GetSettings(c *gin.Context) {
	login := c.GetString("login")
	account := c.Param("account")

	settings, err := cache.GetSettings(c, account, login)
	if err != nil {
		Abort(c, http.StatusNotFound, err, "")
		return
	}

	c.JSON(http.StatusOK, settings)
}

func UpdateSettings(c *gin.Context) {
	login := c.GetString("login")
	account := c.Param("account")

	// check account is associated with login
	// TODO cache this
	installations, err := cache.GetOrCreate[[]string](
		c, "installations", login, 24*time.Hour, func() ([]string, error) {
			return getInstallationAccounts(c, login)
		},
	)
	if err != nil {
		Abort(c, http.StatusInternalServerError, err, "get installations")
		return
	}
	if !lo.Contains(installations, account) {
		Abort(c, http.StatusForbidden, nil, fmt.Sprintf("app not installed to %s, or you have no permission", account))
		return
	}

	var setting cache.Setting
	err = c.ShouldBindJSON(&setting)
	if err != nil {
		Abort(c, http.StatusBadRequest, err, "")
		return
	}

	// check notify settings (check token is valid too)
	_, err = notify.GetNotifier(setting.NotifySettings)
	if err != nil {
		Abort(c, http.StatusBadRequest, err, "invalid notify settings")
		return
	}

	err = cache.SaveSettings(c, account, login, setting)
	if err != nil {
		Abort(c, http.StatusInternalServerError, err, "save settings")
		return
	}

	c.JSON(http.StatusOK, gin.H{})
}

func DeleteSettings(c *gin.Context) {
	login := c.GetString("login")
	account := c.Param("account")

	err := cache.DeleteSettings(c, account, login)
	if err != nil {
		Abort(c, http.StatusInternalServerError, err, "delete settings")
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

func Installations(c *gin.Context) {
	login := c.GetString("login")

	installations, err := getInstallations(c, login)
	if err != nil {
		Abort(c, http.StatusInternalServerError, err, "get installations")
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

	_ = cache.Set(c, "installations", login, accounts, 24*time.Hour)

	c.JSON(http.StatusOK, result)
}

func InstalledRepos(c *gin.Context) {
	installationIDStr := c.Param("installation_id")
	installationID, err := strconv.ParseInt(installationIDStr, 10, 64)
	if err != nil {
		Abort(c, http.StatusBadRequest, nil, "invalid installationID")
		return
	}

	token, err := cache.GetInstallationToken(c, installationID)
	var repoNames []string
	client := github.NewTokenClient(c, token)
	opts := &github.ListOptions{PerPage: 100}
	for {
		repos, resp, err := client.Apps.ListRepos(c, opts)
		if err != nil {
			Abort(c, http.StatusInternalServerError, err, "list repos")
			return
		}
		for _, repo := range repos.Repositories {
			repoNames = append(repoNames, repo.GetFullName())
		}
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	c.JSON(http.StatusOK, repoNames)
}

func CheckSettings(c *gin.Context) {
	var setting cache.Setting
	err := c.ShouldBindJSON(&setting)
	if err != nil {
		Abort(c, http.StatusBadRequest, err, "")
		return
	}

	_, err = notify.GetNotifier(setting.NotifySettings)
	if err != nil {
		Abort(c, http.StatusBadRequest, err, "invalid notify settings")
		return
	}

	c.JSON(http.StatusOK, gin.H{})
}

func TestNotify(c *gin.Context) {
	var setting cache.Setting
	err := c.ShouldBindJSON(&setting)
	if err != nil {
		Abort(c, http.StatusBadRequest, err, "")
		return
	}

	notifier, err := notify.GetNotifier(setting.NotifySettings)
	if err != nil {
		Abort(c, http.StatusBadRequest, err, "invalid notify settings")
		return
	}

	err = notifier.Send(c, "test", "this is a test message")
	if err != nil {
		Abort(c, http.StatusInternalServerError, err, "send test notify")
		return
	}

	c.JSON(http.StatusOK, gin.H{})
}
