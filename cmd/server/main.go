package main

import (
	"net/http"

	"github.com/joho/godotenv"

	"github.com/j178/github_stargazer/api"
	"github.com/j178/github_stargazer/backend/config"
)

func main() {
	_ = godotenv.Load(".env", ".env.local")
	config.Load()

	_ = http.ListenAndServe(":8080", api.Handler())
}
