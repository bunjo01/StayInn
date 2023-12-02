package handlers

import (
	"log"
	"notification/data"
)

type NotificationsHandler struct {
	logger *log.Logger
	repo   *data.NotificationsRepo
}

// Injecting the logger makes this code much more testable
func NewNotificationsHandler(l *log.Logger, r *data.NotificationsRepo) *NotificationsHandler {
	return &NotificationsHandler{l, r}
}

// TODO Handler methods
