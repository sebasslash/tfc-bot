package commands

import (
	"context"

	"github.com/bwmarrin/discordgo"
	embed "github.com/clinet/discordgo-embed"
	"github.com/hashicorp/go-tfe"
	"github.com/sebasslash/tfc-bot/models"
)

var (
	Info chan *models.Message
	Errs chan *models.ErrorMessage
)

type CommandManager struct {
	Ctx      context.Context
	Client   *tfe.Client
	handlers map[string]func(*discordgo.Session, *discordgo.MessageCreate, []string)
}

func (c *CommandManager) Init() {
	c.handlers = make(map[string]func(*discordgo.Session, *discordgo.MessageCreate, []string))
	c.handlers["notify"] = c.notifyHandler
	c.handlers["mute"] = c.muteHandler

	Info = make(chan *models.Message, 100)
	Errs = make(chan *models.ErrorMessage, 10)
}

func (c *CommandManager) ExecuteCmd(s *discordgo.Session, m *discordgo.MessageCreate, command string, args []string) {
	if cmd, ok := c.handlers[command]; ok {
		go cmd(s, m, args)
	}
}

func (c *CommandManager) Listen(s *discordgo.Session) error {
	for {
		select {
		case msg := <-Info:
			s.ChannelMessageSendEmbed(msg.ChannelID, msg.Embed)
		case err := <-Errs:
			s.ChannelMessageSendEmbed(err.ChannelID, embed.NewErrorEmbed(err.Title, err.Err.Error()))
		}
	}
}

func (c *CommandManager) sendErrorMsg(channelID string, err error) {
	Errs <- &models.ErrorMessage{
		Title:     "Oops an error occurred",
		Err:       err,
		ChannelID: channelID,
	}
}
