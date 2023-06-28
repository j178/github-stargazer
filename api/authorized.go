package api

import (
	"net/http"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
)

// 开启 "Request user authorization (OAuth) during installation" 之后，安装的过程同时也是授权的过程
// 用户授权之后，GitHub 会带 code 将用户重定向到这里
// 这里设置 cookie session 后重定向到 setup 页面

func Authorized(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "code is empty", http.StatusBadRequest)
		return
	}

	cfg := oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint:     github.Endpoint,
	}
	token, err := cfg.Exchange(r.Context(), code)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	expire := time.Now().Add(10 * time.Minute)
	cookie := http.Cookie{Name: "token", Value: token.AccessToken, Path: "/", Expires: expire, MaxAge: 90000}
	http.SetCookie(w, &cookie)
	http.Redirect(w, r, "/api/setup", http.StatusFound)
}
