package clients

import (
	"context"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"net/http"
	"reservation/domain"
	"time"

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

func (c AccommodationClient) CheckIfAccommodationExists(ctx context.Context, accommodationId primitive.ObjectID) (interface{}, error) {
	var timeout time.Duration
	deadline, reqHasDeadline := ctx.Deadline()
	if reqHasDeadline {
		timeout = time.Until(deadline)
	}

	cbResp, err := c.cb.Execute(func() (interface{}, error) {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet,
			c.address+"/accommodation/"+accommodationId.Hex(), nil)
		if err != nil {
			return nil, err
		}
		return c.client.Do(req)
	})
	if err != nil {
		handleHttpReqErr(err, c.address+"/accommodation/"+accommodationId.Hex(), http.MethodGet, timeout)
	}

	resp := cbResp.(*http.Response)
	if resp.StatusCode != http.StatusOK {
		return nil, domain.ErrResp{
			URL:        resp.Request.URL.String(),
			Method:     resp.Request.Method,
			StatusCode: resp.StatusCode,
		}
	}

	return true, nil
}
