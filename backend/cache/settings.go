package cache

import (
	"context"
)

type Setting struct {
	NotifySettings []map[string]string `json:"notify_settings"`
	AllowRepos     []string            `json:"allow_repos"`
	MuteRepos      []string            `json:"mute_repos"`
}

func (s *Setting) IsAllowRepo(fullName string) bool {
	for _, repo := range s.MuteRepos {
		if repo == fullName {
			return false
		}
	}
	if len(s.AllowRepos) == 0 {
		return true
	}
	for _, repo := range s.AllowRepos {
		if repo == fullName {
			return true
		}
	}
	return false
}

func GetSettings(ctx context.Context, login string) (*Setting, error) {
	var setting Setting
	err := Get(ctx, "settings", login, &setting)
	if err != nil {
		return nil, err
	}
	return &setting, nil
}

func SaveSettings(ctx context.Context, login string, setting Setting) error {
	return Set(ctx, "settings", login, setting, FOREVER)
}
