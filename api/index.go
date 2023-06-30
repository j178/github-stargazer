package api

import (
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"

	"github.com/j178/github_stargazer/backend/config"
	"github.com/j178/github_stargazer/backend/middleware"
	"github.com/j178/github_stargazer/backend/routes"
)

func InitHandler() http.Handler {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	r.GET(
		"/api/health", func(c *gin.Context) {
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

var (
	handler http.Handler
	once    sync.Once
)

func Index(w http.ResponseWriter, r *http.Request) {
	config.Load()
	once.Do(
		func() {
			handler = InitHandler()
		},
	)
	handler.ServeHTTP(w, r)
}
