package models

import (
	"encoding/json"

	"github.com/hashicorp/go-tfe"
)

type RunNotification struct {
	Message string `json:"message"`
	Trigger string `json:"trigger"`

	Status    tfe.RunStatus `json:"run_status"`
	UpdatedAt string        `json:"run_updated_at"`
	UpdatedBy string        `json:"run_updated_by"`
}

type Notification struct {
	PayloadVersion              int    `json:"payload_version"`
	NotificationConfigurationID string `json:"notification_configuration_id"`

	RunID        string `json:"run_id"`
	RunMessage   string `json:"run_message"`
	RunCreatedAt string `json:"run_created_at"`
	RunCreatedBy string `json:"run_created_by"`

	WorkspaceID   string `json:"workspace_id"`
	WorkspaceName string `json:"workspace_name"`
	Organization  string `json:"organization"`

	RunNotifications []*RunNotification `json:"notifications"`
}

func (n *Notification) MarshalBinary() ([]byte, error) {
	return json.Marshal(n)
}

func (r *RunNotification) MarshalBinary() ([]byte, error) {
	return json.Marshal(r)
}
