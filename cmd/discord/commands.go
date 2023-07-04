package main

import (
	"github.com/bwmarrin/discordgo"
	"github.com/j178/github_stargazer/backend/config"
	"github.com/j178/github_stargazer/backend/notify"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load(".env", ".env.local")
	config.Load()

	bot := notify.DefaultDiscordBot()
	_, err := bot.ApplicationCommandCreate(
		config.DiscordAppID, "", &discordgo.ApplicationCommand{
			Name:        "connect",
			Description: "Connect your GitHub account",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "token",
					Description: "Connect token",
					Required:    true,
				},
			},
		},
	)
	if err != nil {
		panic(err)
	}

	err = bot.Open()
	if err != nil {
		panic(err)
	}
	err = bot.UpdateWatchStatus(0, "Your GitHub Stars")
	if err != nil {
		panic(err)
	}
}
