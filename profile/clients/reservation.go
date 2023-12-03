package clients

import (
	"net/http"

	"github.com/sony/gobreaker"
)

type ReservationClient struct {
	client  *http.Client
	address string
	cb      *gobreaker.CircuitBreaker
}

func NewReservationClient(client *http.Client, address string, cb *gobreaker.CircuitBreaker) ReservationClient {
	return ReservationClient{
		client:  client,
		address: address,
		cb:      cb,
	}
}

// TODO: Client methods
