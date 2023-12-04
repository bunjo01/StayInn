package clients

import (
	"net/http"

	"github.com/sony/gobreaker"
)

type ProfileClient struct {
	client  *http.Client
	address string
	cb      *gobreaker.CircuitBreaker
}

func NewProfileClient(client *http.Client, address string, cb *gobreaker.CircuitBreaker) ProfileClient {
	return ProfileClient{
		client:  client,
		address: address,
		cb:      cb,
	}
}

// TODO: Client methods (checking username and ID)
