package routes

import (
	"context"
	"net/http"

	"github.com/bradleyfalzon/ghinstallation/v2"
	"github.com/gin-gonic/gin"
	"github.com/google/go-github/v53/github"

	"github.com/j178/github_stargazer/config"
)

func GetSettings(c *gin.Context) {
	user := c.GetString("login")
	setting, err := getSettings(user)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, setting)
}

func UpdateSettings(c *gin.Context) {
	user := c.GetString("login")

	var setting Setting
	err := c.ShouldBindJSON(&setting)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = saveSettings(user, setting)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{})
}

func InstalledRepos(c *gin.Context) {
	user := c.GetString("login")
	client, err := getAppClient(c, user)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	opts := &github.ListOptions{PerPage: 100}
	var repoNames []string
	for {
		repos, resp, err := client.Apps.ListRepos(c, opts)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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

func getAppClient(ctx context.Context, login string) (*github.Client, error) {
	// 获取 installationID
	atr, err := ghinstallation.NewAppsTransport(http.DefaultTransport, config.AppID, config.AppPrivateKey)
	if err != nil {
		return nil, err
	}
	client := github.NewClient(&http.Client{Transport: atr})
	installation, _, err := client.Apps.FindUserInstallation(ctx, login)
	if err != nil {
		return nil, err
	}

	// 生成 installation token, 然后获取 installation 有权限的 repo
	tr := ghinstallation.NewFromAppsTransport(atr, installation.GetID())
	client = github.NewClient(&http.Client{Transport: tr})
	return client, nil
}
