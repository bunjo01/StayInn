package handlers

import (
	"log"
	"profile/data"
)

type KeyProduct struct{}

type UserHandler struct {
	logger *log.Logger
	repo   *data.UserRepo
}

// Injecting the logger makes this code much more testable
func NewUserHandler(l *log.Logger, r *data.UserRepo) *UserHandler {
	return &UserHandler{l, r}
}

// TODO Handler methods
