package commands

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/hashicorp/go-tfe"
)

func (c *CommandManager) watchHandler(s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
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

	/* go func() {
		worker := &workers.RunWorker{
			Ctx:           context.Background(),
			Client:        c.Client,
			RunInfo:       RunInfo,
			RunAction:     RunAction,
			RawEmbeds:     RawEmbeds,
			Logs:          OutputLogs,
			Errs:          ErrorMsgs,
			WorkspaceID:   w.ID,
			WorkspaceName: w.Name,
			Organization:  w.Organization.Name,
			ChannelID:     m.ChannelID,
		}

		for {
			err := worker.Run()
			if err != nil {
				c.sendErrorMsg(m.ChannelID, err)
			}
		}
	}() */

	s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Watching workspace **%s** for any runs", w.Name))
}
