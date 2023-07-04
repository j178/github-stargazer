package api

import (
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/j178/github_stargazer/backend/routes/configure"
	"github.com/j178/github_stargazer/backend/routes/discord"
	"github.com/j178/github_stargazer/backend/routes/github"
	"github.com/j178/github_stargazer/backend/routes/telegram"

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

	// Auth
	{
		// redirected from GitHub
		r.GET("/api/authorized", routes.Authorized)
		// redirect to GitHub
		r.GET("/api/authorize", routes.Authorize)
	}
	{
		// GitHub webhook
		r.POST("/api/webhook/github", github.OnEvent)
	}
	// Configure UI API
	{
		checkJWT := middleware.CheckJWT(config.SecretKey)
		admin := r.Group("", checkJWT)
		admin.GET("/api/installations", configure.Installations)
		admin.GET("/api/settings/:account", configure.GetSettings)
		admin.POST("/api/settings/:account", configure.UpdateSettings)
		admin.DELETE("/api/settings/:account", configure.DeleteSettings)
		admin.POST("/api/settings/check", configure.CheckSettings)
		admin.POST("/api/settings/test", configure.TestNotify)
		admin.GET("/api/repos/:installationID", configure.InstalledRepos)
		admin.POST("/api/connect/:platform", configure.GenerateConnectToken)
		admin.GET("/api/connect/:platform/:token", configure.GetConnectResult)
	}
	// Telegram webhook
	{
		r.POST("/api/webhook/telegram", telegram.OnUpdate)
	}
	// Discord interactions endpoint
	{
		r.POST("/api/webhook/discord", discord.OnInteraction)
	}

	return r
}

var Handler = sync.OnceValue(
	func() http.Handler {
		return initRouter().Handler()
	},
)

func Index(w http.ResponseWriter, r *http.Request) {
	config.Load()
	Handler().ServeHTTP(w, r)
}
