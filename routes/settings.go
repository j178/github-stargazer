package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/go-github/v53/github"
	"github.com/j178/github_stargazer/middleware"
)

func GetSettings(c *gin.Context) {
	user := c.MustGet("jwt").(*middleware.JWTClaims).Subject
	setting, err := getSettings(user)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, setting)
}

func UpdateSettings(c *gin.Context) {
	user := c.MustGet("jwt").(*middleware.JWTClaims).Subject

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
	jwt := c.MustGet("jwt").(*middleware.JWTClaims)

	client := github.NewTokenClient(c, jwt.InstallationToken)
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
