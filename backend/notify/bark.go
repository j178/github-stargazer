package notify

import (
	"errors"

	"github.com/nikoksr/notify/service/bark"
)

type barkService struct {
	*bark.Service
}

func (s *barkService) FromSettings(settings map[string]string) error {
	key := settings["key"]
	server := settings["server"]
	if key == "" {
		return errors.New("bark: key is empty")
	}
	if server == "" {
		server = bark.DefaultServerURL
	}

	s.Service = bark.NewWithServers(key, server)
	return nil
}