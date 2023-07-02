package notify

import (
	"context"
	"errors"

	"github.com/bwmarrin/discordgo"
	"github.com/samber/lo"
)

const (
	defaultUsername = "Star++"
	defaultAvatar   = "https://github-stargazer.vercel.app/avatar.png"
)

type discordService struct {
	webhookID    string
	webhookToken string
	params       discordgo.WebhookParams
}

// TODO: add discord bot support

func (s *discordService) Configure(settings map[string]string) error {
	// How to create a discord webhook: https://support.discord.com/hc/en-us/articles/228383668-Intro-to-Webhooks
	webhookID := settings["webhook_id"]
	webhookToken := settings["webhook_token"]
	if webhookID == "" || webhookToken == "" {
		return errors.New("discord: webhook_id or webhook_token is empty")
	}

	s.webhookID = webhookID
	s.webhookToken = webhookToken
	s.params = discordgo.WebhookParams{
		Username:  lo.Ternary(settings["username"] != "", settings["username"], defaultUsername),
		AvatarURL: lo.Ternary(settings["avatar_url"] != "", settings["avatar_url"], defaultAvatar),
	}
	return nil
}

func (s *discordService) Send(ctx context.Context, title, message string) error {
	content := title + "\n" + message
	params := s.params
	params.Content = content

	session, _ := discordgo.New("")
	// https://discord.com/developers/docs/resources/webhook#execute-webhook
	_, err := session.WebhookExecute(s.webhookID, s.webhookToken, false, &params, discordgo.WithContext(ctx))
	if err != nil {
		return err
	}

	return nil
}
