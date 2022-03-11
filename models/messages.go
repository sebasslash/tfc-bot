package models

import (
	"github.com/bwmarrin/discordgo"
	"github.com/hashicorp/go-tfe"
)

type Message struct {
	Embed     *discordgo.MessageEmbed
	ChannelID string
}

type ErrorMessage struct {
	Title     string
	Err       error
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
