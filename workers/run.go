package workers

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	embed "github.com/clinet/discordgo-embed"
	"github.com/hashicorp/go-tfe"
	"github.com/sebasslash/tfc-bot/models"
	"github.com/sebasslash/tfc-bot/store"
)

type RunWorker struct {
	WorkspaceID string
	Client      *tfe.Client
	Info        chan *models.Message
}

func (w *RunWorker) Subscribe(ctx context.Context) error {
	ch, err := store.DB.SubNotificationChannel(ctx, w.WorkspaceID)
	if err != nil {
		return err
	}

	n := &models.Notification{}
	for msg := range ch {
		if err := json.Unmarshal([]byte(msg.Payload), n); err != nil {
			// TODO: handle this better
			panic(err)
		}

		channelID, err := store.DB.GetConfiguration(ctx, n.NotificationConfigurationID)
		if err != nil {
			panic(err)
		}

		if n.RunNotifications[0].Status == tfe.RunCostEstimated {
			if n.RunNotifications[0].Trigger == "run:needs_attention" {
				w.sendNeedsAttentionMessage(channelID, n)
			}

			r, err := w.Client.Runs.Read(ctx, n.RunID)
			if err != nil {
				return err
			}

			w.runCostEstimation(ctx, channelID, r)
		}
		err = w.sendRunNotification(channelID, n)
		if err != nil {
			panic(err)
		}
	}

	return nil
}

func (w *RunWorker) sendRunNotification(channelID string, notification *models.Notification) error {
	pageLink := fmt.Sprintf("%s/app/%s/workspaces/%s/runs/%s",
		os.Getenv("TFE_ADDRESS"),
		notification.Organization,
		notification.WorkspaceName,
		notification.RunID)

	e := embed.NewEmbed().
		SetTitle(notification.RunID).
		SetDescription(notification.RunMessage).
		SetThumbnail("https://www.terraform.io/img/docs/tfe_logo.png").
		SetURL(pageLink).
		AddField("Workspace Name", notification.WorkspaceName).
		AddField("Status", models.PrettyStatus[notification.RunNotifications[0].Status]).
		AddField("Created By", notification.RunCreatedBy).
		InlineAllFields().MessageEmbed

	w.Info <- &models.Message{
		ChannelID: channelID,
		Embed:     e,
	}

	return nil
}

func (w *RunWorker) sendNeedsAttentionMessage(channelID string, notification *models.Notification) error {
	pageLink := fmt.Sprintf("%s/app/%s/workspaces/%s/runs/%s",
		os.Getenv("TFE_ADDRESS"),
		notification.Organization,
		notification.WorkspaceName,
		notification.RunID)

	e := embed.NewEmbed().
		SetTitle(notification.RunID).
		SetDescription("**Awaiting Approval** ⚠️").
		SetThumbnail("https://www.terraform.io/img/docs/tfe_logo.png").
		SetURL(pageLink).
		AddField("Workspace Name", notification.WorkspaceName).
		AddField("Created By", notification.RunCreatedBy).
		InlineAllFields().MessageEmbed

	w.Info <- &models.Message{
		ChannelID: channelID,
		Embed:     e,
	}

	return nil
}

func (w *RunWorker) runCostEstimation(ctx context.Context, channelID string, run *tfe.Run) error {
	ce, err := w.Client.CostEstimates.Read(ctx, run.CostEstimate.ID)
	if err != nil {
		return err
	}

	if ce.Status != tfe.CostEstimateFinished {
		if isErrored(run.Status) || isCanceled(run.Status) {
			return nil
		}
	}

	delta, err := strconv.ParseFloat(ce.DeltaMonthlyCost, 64)
	if err != nil {
		return err
	}

	sign := "+"
	if delta < 0 {
		sign = "-"
	}

	deltaRepr := strings.Replace(ce.DeltaMonthlyCost, "-", "", 1)
	w.sendCostEstimationMsg(channelID, run.ID, ce, sign, deltaRepr)

	return nil
}

func (w *RunWorker) sendCostEstimationMsg(channelID string, runID string, ce *tfe.CostEstimate, sign string, deltaRepr string) {
	e := embed.NewEmbed().
		SetTitle("Cost Estimation for "+runID).
		AddField("Resources", fmt.Sprintf("%d of %d estimated", ce.MatchedResourcesCount, ce.ResourcesCount)).
		AddField("Cost", fmt.Sprintf("$%s/mo %s$%s", ce.ProposedMonthlyCost, sign, deltaRepr)).
		InlineAllFields().MessageEmbed

	w.Info <- &models.Message{
		ChannelID: channelID,
		Embed:     e,
	}
}
