package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/hashicorp/go-tfe"
	"github.com/sebasslash/tfc-bot/commands"
	"github.com/sebasslash/tfc-bot/server"
	"github.com/sebasslash/tfc-bot/store"
)

const (
	CommandPrefix string = "!tfe"
)

var session *discordgo.Session
var cmdManager *commands.CommandManager

func main() {
	ctx := context.Background()

	tfeClient, err := tfe.NewClient(nil)
	if err != nil {
		log.Fatalf("Failed to create TFE client %v", err)
	}

	cmdManager = &commands.CommandManager{
		Client: tfeClient,
		Ctx:    ctx,
	}

	cmdManager.Init()

	store.Create()

	go server.Run(ctx)

	token := fmt.Sprintf("Bot %s", os.Getenv("TFC_BOT_TOKEN"))
	session, err := discordgo.New(token)
	if err != nil {
		log.Fatalf("Error initializing bot: %v", err)
	}

	session.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Println("TFC Bot is running")
		go cmdManager.Listen(s)
	})

	session.AddHandler(handleMessages)
	session.Identify.Intents = discordgo.IntentsGuildMessages

	err = session.Open()
	if err != nil {
		log.Fatalf("Error opening ws connection: %v", err)
	}

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	log.Printf("\r\nShutting down TFC Bot")

	session.Close()
}

func handleMessages(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	if len(m.Content) < len(CommandPrefix) {
		return
	}

	if !hasPrefix(m.Content) {
		return
	}

	command, args := formatCommand(m.Content)
	cmdManager.ExecuteCmd(s, m, command, args)
}

func hasPrefix(msg string) bool {
	return msg[0:4] == CommandPrefix
}

func formatCommand(msg string) (string, []string) {
	// Commands should take the form <prefix> <command> <arguments>
	s := strings.Split(msg, " ")
	switch len(s) {
	case 0, 1:
		return "", []string{}
	case 2:
		return s[1], []string{}
	}

	return s[1], s[2:]
}
