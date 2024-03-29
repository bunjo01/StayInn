package clients

import (
	"context"
	"net/http"
	"profile/domain"
	"time"

	"github.com/sony/gobreaker"
	"go.mongodb.org/mongo-driver/bson/primitive"
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

// Checks reservation service if guest has reservations.
// If guest has reservations, returns error
func (c ReservationClient) CheckUserReservations(ctx context.Context, userID primitive.ObjectID, token string) (interface{}, error) {
	var timeout time.Duration
	deadline, reqHasDeadline := ctx.Deadline()
	if reqHasDeadline {
		timeout = time.Until(deadline)
	}

	cbResp, err := c.cb.Execute(func() (interface{}, error) {
		req, err := http.NewRequestWithContext(ctx, http.MethodDelete,
			c.address+"/user/"+userID.Hex()+"/reservations", nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Authorization", "Bearer "+token)
		return c.client.Do(req)
	})
	if err != nil {
		handleHttpReqErr(err, c.address+"/user/"+userID.Hex()+"/reservations", http.MethodDelete, timeout)
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
