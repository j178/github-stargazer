package cache

import (
	"context"
	"encoding/json"

	"github.com/redis/rueidis"
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

// settings 是两级结构，第一层是 org (如果是个人账号，org 为 login 本身)，第二层是 login

func GetSettings(ctx context.Context, account string, login string) (*Setting, error) {
	redis := Redis()
	cmd := redis.B().Hget().Key(keyFunc("settings", account)).Field(login).Build()
	val, err := redis.Do(ctx, cmd).AsBytes()
	if rueidis.IsRedisNil(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var setting Setting
	err = json.Unmarshal(val, &setting)
	if err != nil {
		return nil, err
	}
	return &setting, nil
}

func GetAllSettings(ctx context.Context, account string) (map[string]*Setting, error) {
	redis := Redis()
	cmd := redis.B().Hgetall().Key(keyFunc("settings", account)).Build()
	val, err := redis.Do(ctx, cmd).AsMap()
	if err != nil {
		return nil, err
	}

	settings := make(map[string]*Setting)
	for k, v := range val {
		var setting Setting
		raw, _ := v.AsBytes()
		err = json.Unmarshal(raw, &setting)
		if err != nil {
			return nil, err
		}
		settings[k] = &setting
	}
	return settings, nil
}

func SaveSettings(ctx context.Context, account, login string, setting Setting) error {
	redis := Redis()
	val, err := json.Marshal(setting)
	if err != nil {
		return err
	}
	cmd := redis.B().Hset().Key(keyFunc("settings", account)).FieldValue().FieldValue(login, string(val)).Build()
	err = redis.Do(ctx, cmd).Error()
	return err
}
