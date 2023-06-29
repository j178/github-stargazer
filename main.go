package main

import (
	"net/http"

	"github.com/joho/godotenv"

	"github.com/j178/github_stargazer/api"
	"github.com/j178/github_stargazer/config"
)

func main() {
	_ = godotenv.Load()
	config.Load()

	_ = http.ListenAndServe(":8080", api.InitHandler())
}
