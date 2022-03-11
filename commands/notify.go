package commands

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/hashicorp/go-tfe"
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
		c.sendErrorMsg(m.ChannelID, fmt.Errorf("Error watching workspace: %v", err))
		return
	}

	go func() {
		nw := &workers.NotificationWorker{
			ChannelID:   m.ChannelID,
			WorkspaceID: w.ID,
			Client:      c.Client,
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

		rw.Subscribe(c.Ctx)
	}()
}
