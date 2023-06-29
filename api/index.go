package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/j178/github_stargazer/api/middleware"
	"github.com/j178/github_stargazer/api/routes"
	"github.com/j178/github_stargazer/config"
)

func initHandler() http.Handler {
	r := gin.Default()

	r.GET(
		"/health", func(c *gin.Context) {
			c.JSON(200, gin.H{"status": "ok"})
		},
	)

	// from GitHub
	{
		r.GET("/api/authorized", routes.Authorized)
		r.POST("/api/hook", routes.OnEvent)
	}
	// from ourselves, needs JWT token
	{
		checkJWT := middleware.CheckJWT([]byte(config.SecretKey))
		r.GET("/api/authorize", checkJWT, routes.Authorize)
		r.GET("/api/settings", checkJWT, routes.GetSettings)
		r.POST("/api/settings", checkJWT, routes.UpdateSettings)
		r.GET("/api/repos", checkJWT, routes.InstalledRepos)
	}

	return r.Handler()
}

var handler = initHandler()

func Index(w http.ResponseWriter, r *http.Request) {
	handler.ServeHTTP(w, r)
}
