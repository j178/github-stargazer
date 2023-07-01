package notify

import (
	"errors"

	"github.com/nikoksr/notify/service/discord"
)

type discordService struct {
	*discord.Discord
}

// TODO: fix this

func (s *discordService) FromSettings(settings map[string]string) error {
	webhook := settings["webhook"]
	if webhook == "" {
		return errors.New("discord: webhook is empty")
	}

	s.Discord = discord.New()
	s.AddReceivers(webhook)
	return nil
}
