package server

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sebasslash/tfc-bot/models"
)

func root(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "TFC Bot Webhook server root")
}

func handleRequests(ctx context.Context) {
	manager := &WebhookManager{
		Notifications: make(chan *models.Notification, 1000),
	}

	go manager.register(ctx)

	r := mux.NewRouter()
	r.HandleFunc("/", root).Methods("GET")
	r.HandleFunc("/run-notifications", manager.handleNotification).Methods("POST")
	http.Handle("/", r)
	log.Fatal(http.ListenAndServe(":10000", nil))
}

func Run(ctx context.Context) {
	log.Printf("Starting webhook server\n")
	handleRequests(ctx)
}
