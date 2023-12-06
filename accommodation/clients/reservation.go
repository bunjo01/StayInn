package clients

import (
	"accommodation/data"
	"accommodation/domain"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

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

func (rc ReservationClient) PassDatesToReservationService(ctx context.Context, accommodationIds []primitive.ObjectID, startDate, endDate time.Time) ([]primitive.ObjectID, error) {
	dates := data.Dates{
		AccommodationIds: accommodationIds,
		StartDate:        startDate,
		EndDate:          endDate,
	}

	requestBody, err := json.Marshal(dates)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal dates: %v", err)
	}

	var timeout time.Duration
	deadline, reqHasDeadline := ctx.Deadline()
	if reqHasDeadline {
		timeout = time.Until(deadline)
	}

	cbResp, err := rc.cb.Execute(func() (interface{}, error) {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, rc.address+"/search", bytes.NewBuffer(requestBody))
		if err != nil {
			return nil, err
		}
		return rc.client.Do(req)
	})
	if err != nil {
		return nil, handleHttpReqErr(err, rc.address+"/search", http.MethodPost, timeout)
	}

	resp := cbResp.(*http.Response)
	if resp.StatusCode != http.StatusOK {
		return nil, domain.ErrResp{
			URL:        resp.Request.URL.String(),
			Method:     resp.Request.Method,
			StatusCode: resp.StatusCode,
		}
	}

	var serviceResponse data.ListOfObjectIds
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&serviceResponse); err != nil {
		return nil, fmt.Errorf("failed to decode JSON response: %v", err)
	}

	return serviceResponse.ObjectIds, nil
}

func (rc ReservationClient) CheckAndDeletePeriods(ctx context.Context, accIDs []primitive.ObjectID) (interface{}, error) {
	requestBody, err := json.Marshal(accIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal accommodation IDs: %v", err)
	}

	var timeout time.Duration
	deadline, reqHasDeadline := ctx.Deadline()
	if reqHasDeadline {
		timeout = time.Until(deadline)
	}

	cbResp, err := rc.cb.Execute(func() (interface{}, error) {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, rc.address+"/check-acc", bytes.NewBuffer(requestBody))
		if err != nil {
			return nil, err
		}
		return rc.client.Do(req)
	})
	if err != nil {
		return nil, handleHttpReqErr(err, rc.address+"/check-acc", http.MethodPost, timeout)
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
