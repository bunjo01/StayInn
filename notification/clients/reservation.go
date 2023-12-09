package clients

import (
	"auth/domain"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reservation/data"
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

func (rc ReservationClient) GetReservationsByUserIDExp(ctx context.Context, accID primitive.ObjectID) (data.Reservations, error) {
	var timeout time.Duration
	deadline, reqHasDeadline := ctx.Deadline()
	if reqHasDeadline {
		timeout = time.Until(deadline)
	}

	url := fmt.Sprintf("%s/reservations/expired/%s", rc.address, accID.Hex())
	cbResp, err := rc.cb.Execute(func() (interface{}, error) {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return nil, err
		}
		resp, err := rc.client.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, domain.ErrResp{
				URL:        resp.Request.URL.String(),
				Method:     resp.Request.Method,
				StatusCode: resp.StatusCode,
			}
		}

		var reservations data.Reservations
		if err := json.NewDecoder(resp.Body).Decode(&reservations); err != nil {
			return nil, err
		}

		return reservations, nil
	})
	if err != nil {
		return nil, handleHttpReqErr(err, url, http.MethodGet, timeout)
	}

	reservations, ok := cbResp.(data.Reservations)
	if !ok {
		return nil, errors.New("invalid response type")
	}

	return reservations, nil
}
