package workers

import (
	"context"
	"fmt"
	"os"

	"github.com/hashicorp/go-tfe"
	"github.com/hashicorp/go-uuid"
	"github.com/sebasslash/tfc-bot/store"
)

type NotificationWorker struct {
	ChannelID   string
	WorkspaceID string

	Client *tfe.Client
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

	fmt.Println("ID: ", nc.ID)

	err = store.DB.AddConfiguration(ctx, nc.ID, w.ChannelID)
	if err != nil {
		return err
	}

	return nil
}

func (w *NotificationWorker) UpdateConfiguration(ctx context.Context, ncID string, enabled bool, triggers *[]string) error {
	opts := tfe.NotificationConfigurationUpdateOptions{
		Enabled: tfe.Bool(enabled),
	}

	if triggers != nil {
		opts.Triggers = *triggers
	}

	_, err := w.Client.NotificationConfigurations.Update(ctx, ncID, opts)
	if err != nil {
		return nil
	}

	return nil
}

func (w *NotificationWorker) DisableConfiguration(ctx context.Context, ncID string) error {
	return w.UpdateConfiguration(ctx, ncID, false, nil)
}
