package main

import (
	"log"
	"net/http"

	"github.com/j178/github_stargazer/api"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/index", api.Index)
	mux.HandleFunc("/api/authorize", api.Authorize)
	mux.HandleFunc("/api/authorized", api.Authorized)
	mux.HandleFunc("/api/setup", api.Setup)
	mux.HandleFunc("/api/hook", api.OnEvent)

	log.Fatal(http.ListenAndServe(":8080", mux))
}
