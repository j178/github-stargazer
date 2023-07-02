package notify

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/bwmarrin/discordgo"
	"github.com/j178/github_stargazer/backend/utils"
)

const (
	authorUrl       = "https://github.com/apps/stars-notifier"
	defaultUsername = "Star++"
	defaultAvatar   = "https://github-stargazer.vercel.app/avatar.png"
	// https://www.spycolor.com/fd9a00
	defaultColor = "fd9a00"
)

type discordService struct {
	webhookID    string
	webhookToken string
	username     string
	avatarURL    string
	color        int64
}

// TODO: add discord bot support

func (s *discordService) Configure(settings map[string]string) error {
	// How to create a discord webhook: https://support.discord.com/hc/en-us/articles/228383668-Intro-to-Webhooks
	webhookID := settings["webhook_id"]
	webhookToken := settings["webhook_token"]
	if webhookID == "" || webhookToken == "" {
		return errors.New("webhook_id or webhook_token is empty")
	}

	s.webhookID = webhookID
	s.webhookToken = webhookToken
	s.username = utils.Or(settings["username"], defaultUsername)
	s.avatarURL = utils.Or(settings["avatar_url"], defaultAvatar)

	color := utils.Or(settings["color"], defaultColor)
	var err error
	s.color, err = strconv.ParseInt(color, 16, 32)
	if err != nil {
		return fmt.Errorf("invalid color")
	}
	return nil
}

func (s *discordService) Send(ctx context.Context, title, message string) error {
	params := discordgo.WebhookParams{}
	params.Embeds = []*discordgo.MessageEmbed{
		{
			Author: &discordgo.MessageEmbedAuthor{
				Name:    s.username,
				IconURL: s.avatarURL,
				URL:     authorUrl,
			},
			Color:       int(s.color),
			Title:       title,
			Description: message,
		},
	}

	session, _ := discordgo.New("")
	// https://discord.com/developers/docs/resources/webhook#execute-webhook
	_, err := session.WebhookExecute(s.webhookID, s.webhookToken, false, &params, discordgo.WithContext(ctx))
	if err != nil {
		return fmt.Errorf("discord webhook: %w", err)
	}

	return nil
}
