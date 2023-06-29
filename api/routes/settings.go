package routes

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/go-github/v53/github"
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
	// TODO use installation token
	accessToken := c.GetString("access_token")

	client := github.NewTokenClient(context.Background(), accessToken)
	opts := &github.ListOptions{PerPage: 100}
	repos, resp, err := client.Apps.ListUserRepos(context.Background(), 1, opts)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	repoNames := make([]string, 0, repos.GetTotalCount())
	for {
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
		repos, resp, err = client.Apps.ListUserRepos(context.Background(), 1, opts)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		for _, repo := range repos.Repositories {
			repoNames = append(repoNames, repo.GetFullName())
		}
	}

	c.JSON(http.StatusOK, repoNames)
}
