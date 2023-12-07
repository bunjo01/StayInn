package clients

import (
	"context"
	"net/http"
	"profile/domain"
	"time"

	"github.com/sony/gobreaker"
	"go.mongodb.org/mongo-driver/bson/primitive"
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

// TODO: Client methods

// Checks accommodation service if host has active reservations on his accommodation
// If host has no reservations, deletes accommodations and it's periods
// Otherwise returns erorr
func (c AccommodationClient) CheckAndDeleteUserAccommodations(ctx context.Context, userID primitive.ObjectID) (interface{}, error) {
	var timeout time.Duration
	deadline, reqHasDeadline := ctx.Deadline()
	if reqHasDeadline {
		timeout = time.Until(deadline)
	}

	cbResp, err := c.cb.Execute(func() (interface{}, error) {
		req, err := http.NewRequestWithContext(ctx, http.MethodDelete,
			c.address+"/user/"+userID.Hex()+"/accommodations", nil)
		if err != nil {
			return nil, err
		}
		return c.client.Do(req)
	})
	if err != nil {
		handleHttpReqErr(err, c.address+"/user/"+userID.Hex()+"/accommodations", http.MethodDelete, timeout)
	}

	resp := cbResp.(*http.Response)
	if resp.StatusCode != http.StatusNoContent {
		return nil, domain.ErrResp{
			URL:        resp.Request.URL.String(),
			Method:     resp.Request.Method,
			StatusCode: resp.StatusCode,
		}
	}

	return true, nil
}
