package workers

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	embed "github.com/clinet/discordgo-embed"
	"github.com/sebasslash/tfc-bot/models"
	"github.com/sebasslash/tfc-bot/store"
)

type RunWorker struct {
	WorkspaceID string
	Info        chan *models.EmbedMessage
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
		AddField("Workspace ID", notification.WorkspaceID).
		AddField("Status", models.PrettyStatus[notification.RunNotifications[0].Status]).
		AddField("Created By", notification.RunCreatedBy).
		InlineAllFields().MessageEmbed

	w.Info <- &models.EmbedMessage{
		ChannelID: channelID,
		Embed:     e,
	}

	return nil
}
