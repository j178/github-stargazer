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
		c.AbortWithStatusJSON(http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, setting)
}

func UpdateSettings(c *gin.Context) {
	user := c.GetString("login")

	var setting Setting
	err := c.ShouldBindJSON(&setting)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, err.Error())
		return
	}

	err = saveSettings(user, setting)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{})
}

func InstalledRepos(c *gin.Context) {
	accessToken := c.GetString("access_token")

	client := github.NewTokenClient(context.Background(), accessToken)
	repos, _, err := client.Apps.ListUserRepos(context.Background(), 1, nil)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, err.Error())
		return
	}
	repoNames := make([]string, len(repos.Repositories))
	for i, repo := range repos.Repositories {
		repoNames[i] = repo.GetFullName()
	}
	c.JSON(http.StatusOK, repoNames)
}
