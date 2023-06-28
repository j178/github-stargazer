package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

func Setup(w http.ResponseWriter, r *http.Request) {
	token, err := r.Cookie("token")
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	user, err := getUser(token.Value)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	if r.Method == http.MethodGet {
		w.WriteHeader(http.StatusOK)
		// TODO configure page
		// TODO check installation
		w.Write([]byte(fmt.Sprintf("<p>Hello, %s</p><i>This is a configuration page.</i>", user)))
		return
	}

	var setting Setting
	err = json.NewDecoder(r.Body).Decode(&setting)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = saveSettings(user, setting)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	// TODO flash "saved"
	w.Write([]byte("Saved"))
	return
}

func getUser(token string) (string, error) {
	req, _ := http.NewRequest("GET", "https://api.github.com/user", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", errors.New(string(body))
	}

	var user struct {
		Login string `json:"login"`
	}
	err = json.Unmarshal(body, &user)
	if err != nil {
		return "", err
	}
	return user.Login, nil
}
