package api

import (
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"

	"github.com/j178/github_stargazer/backend/config"
	"github.com/j178/github_stargazer/backend/middleware"
	"github.com/j178/github_stargazer/backend/routes"
)

func initRouter() *gin.Engine {
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
	// from ourselves, may need JWT token
	r.GET("/api/authorize", routes.Authorize)
	{
		checkJWT := middleware.CheckJWT(config.SecretKey)
		r.GET("/api/installations", checkJWT, routes.Installations)
		r.GET("/api/:account/settings", checkJWT, routes.GetSettings)
		r.POST("/api/:account/settings", checkJWT, routes.UpdateSettings)
		r.GET("/api/repos", checkJWT, routes.InstalledRepos)
	}

	return r
}

var (
	handler http.Handler
	once    sync.Once
)

func Handler() http.Handler {
	once.Do(
		func() {
			handler = initRouter().Handler()
		},
	)
	return handler
}

func Index(w http.ResponseWriter, r *http.Request) {
	config.Load()
	Handler().ServeHTTP(w, r)
}
