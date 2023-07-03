package notify

import (
	"errors"

	"github.com/nikoksr/notify/service/discord"

	"github.com/j178/github_stargazer/backend/config"
)

type discordBotService struct {
	*discord.Discord
}

func (d *discordBotService) Name() string {
	return "discord_bot"
}

func (d *discordBotService) Configure(settings map[string]string) error {
	token := settings["token"]
	channelID := settings["channel_id"]
	if channelID == "" {
		return errors.New("token or channel_id is empty")
	}
	if token == "" || token == "default" {
		token = config.DiscordBotToken
	}

	disc := discord.New()
	err := disc.AuthenticateWithBotToken(token)
	if err != nil {
		return err
	}
	disc.AddReceivers(channelID)
	d.Discord = disc

	return nil
}
