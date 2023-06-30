package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/go-github/v53/github"

	"github.com/j178/github_stargazer/backend/cache"
)

func GetSettings(c *gin.Context) {
	login := c.GetString("login")
	setting, err := cache.GetSettings(c, login)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, setting)
}

func UpdateSettings(c *gin.Context) {
	login := c.GetString("login")

	var setting cache.Setting
	err := c.ShouldBindJSON(&setting)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = cache.SaveSettings(c, login, setting)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{})
}

func InstalledRepos(c *gin.Context) {
	login := c.GetString("login")

	installationToken, err := cache.GetInstallationToken(c, login)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	client := github.NewTokenClient(c, installationToken)
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
