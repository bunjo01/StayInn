package clients

import (
	"net/http"

	"github.com/sony/gobreaker"
)

type AccommodationClient struct {
	client  *http.Client
	address string
	cb      *gobreaker.CircuitBreaker
}

func NewAccommodationClient(client *http.Client, address string, cb *gobreaker.CircuitBreaker) AccommodationClient {
	return AccommodationClient{
		client:  client,
		address: address,
		cb:      cb,
	}
}

// TODO: Client methods (checking accommodation ID when creating period or reservation)
