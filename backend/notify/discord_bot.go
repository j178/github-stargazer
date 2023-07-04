package notify

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strconv"
	"sync"

	"github.com/bwmarrin/discordgo"
	"github.com/j178/github_stargazer/backend/config"
	"github.com/j178/github_stargazer/backend/utils"
)

var DefaultDiscordBot = sync.OnceValue(
	func() *discordgo.Session {
		defaultDiscordBot, err := discordgo.New("Bot " + config.DiscordBotToken)
		if err != nil {
			log.Fatal("init discord bot: %w", err)
		}
		return defaultDiscordBot
	},
)

type discordBotService struct {
	bot       *discordgo.Session
	channelID string
	username  string
	avatarURL string
	color     int64
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

	d.channelID = channelID
	d.username = utils.Or(settings["username"], defaultUsername)
	d.avatarURL = utils.Or(settings["avatar_url"], defaultAvatar)
	color := utils.Or(settings["color"], defaultColor)
	var err error
	d.color, err = strconv.ParseInt(color, 16, 32)
	if err != nil {
		return fmt.Errorf("invalid color")
	}

	var bot *discordgo.Session
	if token == "" || token == "default" {
		bot = DefaultDiscordBot()
	} else {
		bot, err = discordgo.New("Bot " + token)
		if err != nil {
			return err
		}
	}
	d.bot = bot

	return nil
}

func (d *discordBotService) Send(ctx context.Context, title, message string) error {
	embed := &discordgo.MessageEmbed{
		Author: &discordgo.MessageEmbedAuthor{
			Name:    d.username,
			IconURL: d.avatarURL,
			URL:     authorUrl,
		},
		Color:       int(d.color),
		Title:       title,
		Description: message,
	}
	_, err := d.bot.ChannelMessageSendEmbed(d.channelID, embed, discordgo.WithContext(ctx))
	if err != nil {
		return fmt.Errorf("discord bot send message: %w", err)
	}
	return nil
}
