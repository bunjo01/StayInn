package handlers

import (
	"auth/data"
	"log"
)

type KeyProduct struct{}

type CredentialsHandler struct {
	logger *log.Logger
	repo   *data.CredentialsRepo
}

// Injecting the logger makes this code much more testable
func NewCredentialsHandler(l *log.Logger, r *data.CredentialsRepo) *CredentialsHandler {
	return &CredentialsHandler{l, r}
}

// TODO Handler methods
