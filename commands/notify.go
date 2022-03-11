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

func (c *CommandManager) notifyHandler(s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	if len(args) == 0 {
		return
	}

	w, err := c.Client.Workspaces.ReadByIDWithOptions(c.Ctx, args[0], &tfe.WorkspaceReadOptions{
		Include: "organization",
	})
	if err != nil {
		c.sendErrorMsg(m.ChannelID, fmt.Errorf("Error fetching workspace: %v", err))
		return
	}

	key, err := store.DB.ReadConfigurationKey(c.Ctx, &models.ConfigurationKey{
		WorkspaceID: w.ID,
		ChannelID:   m.ChannelID,
	})
	if err != nil && err != redis.Nil {
		c.sendErrorMsg(m.ChannelID, err)
		return
	}

	if key != "" {
		c.sendErrorMsg(m.ChannelID, fmt.Errorf("Notification configuration for workspace %s aleady exists for this channel.", w.ID))
		return
	}

	go func() {
		nw := &workers.NotificationWorker{
			ChannelID:     m.ChannelID,
			WorkspaceID:   w.ID,
			WorkspaceName: w.Name,
			Organization:  w.Organization.Name,
			Client:        c.Client,

			Info: Info,
		}

		err := nw.CreateConfiguration(c.Ctx)
		if err != nil {
			c.sendErrorMsg(m.ChannelID, err)
		}
	}()

	go func() {
		rw := &workers.RunWorker{
			WorkspaceID: w.ID,
			Info:        Info,
		}

		err := rw.Subscribe(c.Ctx)
		if err != nil {
			c.sendErrorMsg(m.ChannelID, err)
		}
	}()
}
