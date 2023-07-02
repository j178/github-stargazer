package notify

import (
	"errors"

	"github.com/nikoksr/notify/service/bark"
)

type barkService struct {
	*bark.Service
}

func (s *barkService) Configure(settings map[string]string) error {
	key := settings["device_token"]
	server := settings["server"]
	if key == "" {
		return errors.New("bark: device_token is empty")
	}
	if server == "" {
		server = bark.DefaultServerURL
	}

	s.Service = bark.NewWithServers(key, server)
	return nil
}
