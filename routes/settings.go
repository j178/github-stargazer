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

	appTransport, err := ghinstallation.NewAppsTransport(http.DefaultTransport, config.AppID, config.AppPrivateKey)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	client := github.NewClient(&http.Client{Transport: appTransport})
	installation, _, err := client.Apps.FindUserInstallation(context.Background(), user)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	installToken, _, err := client.Apps.CreateInstallationToken(context.Background(), installation.GetID(), nil)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	repoNames := make([]string, len(installToken.Repositories))
	for i, repo := range installToken.Repositories {
		repoNames[i] = repo.GetFullName()
	}

	c.JSON(http.StatusOK, repoNames)
}
