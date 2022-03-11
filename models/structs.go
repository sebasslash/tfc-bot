package models

import (
	"os"

	"github.com/bwmarrin/discordgo"
	"github.com/hashicorp/go-tfe"
)

type RunStatus struct {
	Title     string
	ChannelID string
	Run       *tfe.Run
	PageLink  string
}

type RunActionMessage struct {
	RunID     string
	ChannelID string
	Status    tfe.RunStatus
	Success   bool
}

type StatusMessage struct {
	Title     string
	Message   string
	ChannelID string
}

type LogMessage struct {
	Title     string
	File      *os.File
	ChannelID string
}

type EmbedMessage struct {
	Embed     *discordgo.MessageEmbed
	ChannelID string
}

var PrettyStatus map[tfe.RunStatus]string = map[tfe.RunStatus]string{
	tfe.RunPending:        "Run Pending ⏸",
	tfe.RunPlanning:       "Planning 🔁",
	tfe.RunPlanned:        "Plan successful ✅",
	tfe.RunPlanQueued:     "Plan Queued ⏸",
	tfe.RunCostEstimating: "Cost Estimating 💰",
	tfe.RunDiscarded:      "Run Discarded 🗑️",
	tfe.RunErrored:        "An error occurred ❌",
	tfe.RunConfirmed:      "Run Confirmed ✅",
	tfe.RunApplyQueued:    "Apply Queued ⏸",
	tfe.RunApplying:       "Applying 🔁",
	tfe.RunApplied:        "Apply successful ✅",
}
