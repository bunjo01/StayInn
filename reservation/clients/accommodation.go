package clients

import (
	"context"
	"net/http"
	"reservation/domain"
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

func (ac *AccommodationClient) CheckAccommodationID(ctx context.Context, accID primitive.ObjectID, token string) (bool, error) {
	var timeout time.Duration
	deadline, reqHasDeadline := ctx.Deadline()
	if reqHasDeadline {
		timeout = time.Until(deadline)
	}

	cbResp, err := ac.cb.Execute(func() (interface{}, error) {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, ac.address+"/accommodation/"+accID.Hex(), nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Authorization", "Bearer "+token)
		return ac.client.Do(req)
	})
	if err != nil {
		return false, handleHttpReqErr(err, ac.address+"/accommodation/"+accID.Hex(), http.MethodGet, timeout)
	}

	resp := cbResp.(*http.Response)
	if resp.StatusCode != http.StatusOK {
		return false, domain.ErrResp{
			URL:        resp.Request.URL.String(),
			Method:     resp.Request.Method,
			StatusCode: resp.StatusCode,
		}
	}

	return true, nil
}
