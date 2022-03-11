package server

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/sebasslash/tfc-bot/models"
	"github.com/sebasslash/tfc-bot/store"
)

type WebhookManager struct {
	Notifications chan *models.Notification
}

func (m *WebhookManager) register(ctx context.Context) {
	for n := range m.Notifications {
		err := store.DB.PublishNotification(ctx, n)
		if err != nil {
			log.Printf("[ERROR] %v\n", err)
		}
	}
}

func (m *WebhookManager) handleNotification(w http.ResponseWriter, r *http.Request) {
	n, err := m.parseNotification(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	m.Notifications <- n

	w.WriteHeader(http.StatusAccepted)
}

func (m *WebhookManager) parseNotification(r *http.Request) (*models.Notification, error) {
	var n models.Notification
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(body, &n)
	if err != nil {
		return nil, err
	}

	return &n, nil
}
