package clients

import (
	"net/http"

	"github.com/sony/gobreaker"
)

type NotificationClient struct {
	client  *http.Client
	address string
	cb      *gobreaker.CircuitBreaker
}

func NewNotificationClient(client *http.Client, address string, cb *gobreaker.CircuitBreaker) NotificationClient {
	return NotificationClient{
		client:  client,
		address: address,
		cb:      cb,
	}
}

// TODO: Client methods (notify notification service when reservation is created or deleted)
