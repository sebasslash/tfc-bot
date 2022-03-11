package workers

import (
	"context"
	"fmt"
	"os"

	embed "github.com/clinet/discordgo-embed"
	"github.com/hashicorp/go-tfe"
	"github.com/hashicorp/go-uuid"
	"github.com/sebasslash/tfc-bot/models"
	"github.com/sebasslash/tfc-bot/store"
)

type Action string

const (
	Created  Action = "created"
	Updated  Action = "updated"
	Disabled Action = "disabled"
)

type NotificationWorker struct {
	ChannelID     string
	WorkspaceID   string
	WorkspaceName string
	Organization  string

	Client *tfe.Client
	Info   chan *models.Message
}

func (w *NotificationWorker) CreateConfiguration(ctx context.Context) error {
	name, err := uuid.GenerateUUID()
	if err != nil {
		return err
	}

	opts := tfe.NotificationConfigurationCreateOptions{
		DestinationType: tfe.NotificationDestination(tfe.NotificationDestinationTypeGeneric),
		Enabled:         tfe.Bool(true),
		Name:            tfe.String(fmt.Sprintf("tfc-bot-%s", name)),
		Triggers: []string{
			tfe.NotificationTriggerCreated,
			tfe.NotificationTriggerPlanning,
			tfe.NotificationTriggerNeedsAttention,
			tfe.NotificationTriggerApplying,
			tfe.NotificationTriggerCompleted,
			tfe.NotificationTriggerErrored,
		},
		URL: tfe.String(os.Getenv("TFC_BOT_WEBHOOK_URL")),
	}

	nc, err := w.Client.NotificationConfigurations.Create(ctx, w.WorkspaceID, opts)
	if err != nil {
		return err
	}

	err = store.DB.CreateConfigurationKey(ctx, &models.ConfigurationKey{
		WorkspaceID: w.WorkspaceID,
		ChannelID:   w.ChannelID,
	}, nc.ID)
	if err != nil {
		return err
	}

	err = store.DB.AddConfiguration(ctx, nc.ID, w.ChannelID)
	if err != nil {
		return err
	}

	w.sendConfigurationUpdate(ctx, nc.ID, Created)

	return nil
}

func (w *NotificationWorker) UpdateConfiguration(ctx context.Context, ncID string, enabled bool, triggers *[]string) error {
	opts := tfe.NotificationConfigurationUpdateOptions{Enabled: tfe.Bool(enabled)}

	if triggers != nil {
		opts.Triggers = *triggers
	}

	nc, err := w.Client.NotificationConfigurations.Update(ctx, ncID, opts)
	if err != nil {
		return nil
	}

	if !nc.Enabled {
		w.sendConfigurationUpdate(ctx, nc.ID, Disabled)
	} else {
		w.sendConfigurationUpdate(ctx, nc.ID, Updated)
	}

	return nil
}

func (w *NotificationWorker) DisableConfiguration(ctx context.Context, ncID string) error {
	return w.UpdateConfiguration(ctx, ncID, false, nil)
}

func (w *NotificationWorker) sendConfigurationUpdate(ctx context.Context, ncID string, action Action) {
	desc := fmt.Sprintf("Notification configuration (%s) %s for workspace %s", ncID, string(action), w.WorkspaceID)
	link := fmt.Sprintf("%s/app/%s/workspaces/%s/settings/notifications/%s",
		os.Getenv("TFE_ADDRESS"),
		w.Organization,
		w.WorkspaceName,
		ncID,
	)

	e := embed.NewEmbed().
		SetTitle(fmt.Sprintf("%s (workspace %s)", ncID, w.WorkspaceID)).
		SetDescription(desc).
		SetURL(link)

	switch action {
	case Created:
		e.SetColor(0x00ff00)
	case Updated:
		e.SetColor(0x0000ff)
	case Disabled:
		e.SetColor(0xff0000)
	}

	w.Info <- &models.Message{
		ChannelID: w.ChannelID,
		Embed:     e.MessageEmbed,
	}
}
