package commands

import (
	"context"
	"fmt"
	"log"
	"strconv"

	"github.com/bwmarrin/discordgo"
	embed "github.com/clinet/discordgo-embed"
	"github.com/hashicorp/go-tfe"
	"github.com/sebasslash/tfc-bot/models"
)

var (
	RunInfo    chan *models.RunStatus
	RunAction  chan *models.RunActionMessage
	OutputLogs chan *models.LogMessage
	ErrorMsgs  chan *models.StatusMessage
	RawEmbeds  chan *models.EmbedMessage
)

type CommandManager struct {
	Ctx      context.Context
	Client   *tfe.Client
	handlers map[string]func(*discordgo.Session, *discordgo.MessageCreate, []string)
}

func (c *CommandManager) Init() {
	c.handlers = make(map[string]func(*discordgo.Session, *discordgo.MessageCreate, []string))
	c.handlers["watch"] = c.watchHandler
	c.handlers["notify"] = c.notifyHandler

	RunInfo = make(chan *models.RunStatus, 100)
	RunAction = make(chan *models.RunActionMessage, 100)
	OutputLogs = make(chan *models.LogMessage, 100)
	ErrorMsgs = make(chan *models.StatusMessage, 10)
	RawEmbeds = make(chan *models.EmbedMessage, 10)
}

func (c *CommandManager) ExecuteCmd(s *discordgo.Session, m *discordgo.MessageCreate, command string, args []string) {
	if cmd, ok := c.handlers[command]; ok {
		go cmd(s, m, args)
	}
}

func (c *CommandManager) Listen(s *discordgo.Session) error {
	for {
		select {
		case msg := <-RunInfo:
			e := embed.NewEmbed().
				SetTitle(msg.Title).
				SetDescription(msg.Run.Message).
				SetThumbnail("https://www.terraform.io/img/docs/tfe_logo.png").
				SetURL(msg.PageLink).
				AddField("Workspace ID", msg.Run.Workspace.ID).
				AddField("Status", models.PrettyStatus[msg.Run.Status]).
				AddField("Auto-Apply", strconv.FormatBool(msg.Run.AutoApply)).
				AddField("Source", fmt.Sprintf("%s (%s)", msg.Run.Message, string(msg.Run.Source))).
				InlineAllFields().MessageEmbed

			_, err := s.ChannelMessageSendEmbed(msg.ChannelID, e)
			if err != nil {
				fmt.Println(err)
			}
		case embed := <-RawEmbeds:
			s.ChannelMessageSendEmbed(embed.ChannelID, embed.Embed)
		case msg := <-RunAction:
			e := embed.NewEmbed().SetTitle(msg.RunID).SetDescription(models.PrettyStatus[msg.Status])
			if msg.Success {
				e.SetColor(0x00ff00)
			} else {
				e.SetColor(0xb40000)
			}
			s.ChannelMessageSendEmbed(msg.ChannelID, e.MessageEmbed)
		case msg := <-OutputLogs:
			ms := &discordgo.MessageSend{
				/* Embed: &discordgo.MessageEmbed{
					Title: "See Plan Output",
					Image: &discordgo.MessageEmbedImage{
						URL: "attachment://" + msg.File.Name(),
					},
				}, */
				Files: []*discordgo.File{
					{
						Name:   msg.File.Name(),
						Reader: msg.File,
					},
				},
			}

			_, err := s.ChannelMessageSendComplex(msg.ChannelID, ms)
			// _, err = s.ChannelFileSend(msg.ChannelID, msg.Filename, file)
			if err != nil {
				log.Printf("[ERROR] %v", err)
			}
		case err := <-ErrorMsgs:
			s.ChannelMessageSendEmbed(err.ChannelID, embed.NewErrorEmbed(err.Title, err.Message))
		}
	}
}

func (c *CommandManager) sendErrorMsg(channelID string, err error) {
	ErrorMsgs <- &models.StatusMessage{
		Title:     "Oops an error",
		Message:   err.Error(),
		ChannelID: channelID,
	}
}
