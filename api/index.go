package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/j178/github_stargazer/config"
	"github.com/j178/github_stargazer/middleware"
	"github.com/j178/github_stargazer/routes"
)

func initHandler() http.Handler {
	gin.SetMode(gin.ReleaseMode)
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
		checkJWT := middleware.CheckJWT(config.SecretKey)
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
