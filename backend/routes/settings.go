package routes

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/go-github/v53/github"

	"github.com/j178/github_stargazer/backend/cache"
)

func GetSettings(c *gin.Context) {
	login := c.Query("login")
	account := c.Param("account")

	settings, err := cache.GetSettings(c, account, login)
	if err != nil {
		Abort(c, http.StatusNotFound, err, "")
		return
	}

	c.JSON(http.StatusOK, settings)
}

func UpdateSettings(c *gin.Context) {
	login := c.Query("login")
	account := c.Param("account")

	var setting cache.Setting
	err := c.ShouldBindJSON(&setting)
	if err != nil {
		Abort(c, http.StatusBadRequest, err, "")
		return
	}

	err = cache.SaveSettings(c, account, login, setting)
	if err != nil {
		Abort(c, http.StatusInternalServerError, err, "save settings")
		return
	}

	c.JSON(http.StatusOK, gin.H{})
}

type Installation struct {
	ID          int64  `json:"id"`
	Account     string `json:"account"`
	AccountType string `json:"account_type"`
}

func Installations(c *gin.Context) {
	login := c.GetString("login")

	token, err := cache.GetOAuthToken(c, login)
	if err != nil {
		Abort(c, http.StatusInternalServerError, err, "get access token")
		return
	}

	client := github.NewTokenClient(c, token)
	opts := &github.ListOptions{PerPage: 100}

	var installations []Installation
	for {
		ins, resp, err := client.Apps.ListUserInstallations(c, opts)
		if err != nil {
			Abort(c, http.StatusInternalServerError, err, "list installations")
			return
		}
		for _, i := range ins {
			installations = append(
				installations,
				Installation{
					ID:          i.GetID(),
					Account:     i.Account.GetLogin(),
					AccountType: i.Account.GetType(),
				},
			)
		}
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	c.JSON(http.StatusOK, installations)
}

func InstalledRepos(c *gin.Context) {
	installationIDStr := c.Query("installation_id")
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
