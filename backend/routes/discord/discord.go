package discord

import (
	"fmt"
	"log"
	"net/http"

	"github.com/bwmarrin/discordgo"
	"github.com/gin-gonic/gin"
	"github.com/j178/github_stargazer/backend/config"
	"github.com/j178/github_stargazer/backend/notify"
	"github.com/j178/github_stargazer/backend/routes/configure"
)

func Bot() *discordgo.Session {
	return notify.DefaultDiscordBot()
}

func OnInteraction(c *gin.Context) {
	if !discordgo.VerifyInteraction(c.Request, config.DiscordPublicKey) {
		c.JSON(http.StatusBadRequest, gin.H{"status": "invalid interaction"})
		return
	}

	var interaction discordgo.Interaction
	err := c.ShouldBindJSON(&interaction)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "invalid interaction"})
		return
	}

	log.Printf("discord interaction: %+v", interaction)

	switch interaction.Type {
	case discordgo.InteractionPing:
		c.JSON(http.StatusOK, discordgo.InteractionResponse{Type: discordgo.InteractionResponsePong})
		return
	case discordgo.InteractionApplicationCommand:
		reply := discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{},
		}
		data := interaction.ApplicationCommandData()
		token := ""
		if len(data.Options) > 0 {
			token = data.Options[0].StringValue()
		}
		if token == "" {
			reply.Data.Content = "Use `/connect` command to connect your GitHub account"
			c.JSON(http.StatusOK, reply)
			return
		}

		connect := map[string]any{
			"guild_id":   interaction.GuildID,
			"channel_id": interaction.ChannelID,
		}
		err := configure.SetConnectResult(c, token, "discord", connect)
		if err != nil {
			reply.Data.Content = fmt.Sprintf("Invalid connect token: %q", token)
			c.JSON(http.StatusOK, reply)
			return
		}

		reply.Data.Content = "Connected!"
		c.JSON(http.StatusOK, reply)
		return
	}
}
