package clients

import (
	"accommodation/data"
	"accommodation/domain"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"net/http"
	"time"

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

// TODO: Client methods (search/filter according to start and end date of travel)
func (rc ReservationClient) PassDatesToReservationService(ctx context.Context, startDate, endDate time.Time) ([]primitive.ObjectID, error) {
	dates := data.Dates{
		StartDate: startDate,
		EndDate:   endDate,
	}

	println("38")

	requestBody, err := json.Marshal(dates)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal dates: %v", err)
	}

	println("45")

	var timeout time.Duration
	deadline, reqHasDeadline := ctx.Deadline()
	if reqHasDeadline {
		timeout = time.Until(deadline)
	}

	println("53")

	cbResp, err := rc.cb.Execute(func() (interface{}, error) {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, rc.address+"/search", bytes.NewBuffer(requestBody))
		if err != nil {
			println("58")
			return nil, err
		}
		return rc.client.Do(req)
	})
	println("Response ", &cbResp)
	println("Error ", &err)
	if err != nil {
		println("64")
		return nil, handleHttpReqErr(err, rc.address+"/search", http.MethodPost, timeout)
	}

	println("68")

	resp := cbResp.(*http.Response)
	if resp.StatusCode != http.StatusOK {
		return nil, domain.ErrResp{
			URL:        resp.Request.URL.String(),
			Method:     resp.Request.Method,
			StatusCode: resp.StatusCode,
		}
	}

	println("79")

	// Parse the JSON response
	var serviceResponse data.ListOfObjectIds
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&serviceResponse); err != nil {
		return nil, fmt.Errorf("failed to decode JSON response: %v", err)
	}

	println("88")

	return serviceResponse.ObjectIds, nil
}
