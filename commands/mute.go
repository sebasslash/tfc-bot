package commands

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/go-redis/redis/v8"
	"github.com/hashicorp/go-tfe"
	"github.com/sebasslash/tfc-bot/models"
	"github.com/sebasslash/tfc-bot/store"
	"github.com/sebasslash/tfc-bot/workers"
)

func (c *CommandManager) muteHandler(s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	if len(args) == 0 {
		return
	}

	w, err := c.Client.Workspaces.ReadByIDWithOptions(c.Ctx, args[0], &tfe.WorkspaceReadOptions{
		Include: "organization",
	})
	if err != nil {
		c.sendErrorMsg(m.ChannelID, fmt.Errorf("Error watching workspace: %v", err))
		return
	}

	ncID, err := store.DB.ReadConfigurationKey(c.Ctx, &models.ConfigurationKey{
		WorkspaceID: w.ID,
		ChannelID:   m.ChannelID,
	})
	if err == redis.Nil {
		c.sendErrorMsg(m.ChannelID, fmt.Errorf("No notification configuration found for workspace %s in this channel", w.ID))
	} else if err != nil {
		c.sendErrorMsg(m.ChannelID, err)
	}

	go func() {
		nw := &workers.NotificationWorker{
			ChannelID:   m.ChannelID,
			WorkspaceID: w.ID,
			Client:      c.Client,
		}

		nw.DisableConfiguration(c.Ctx, ncID)
	}()
}
