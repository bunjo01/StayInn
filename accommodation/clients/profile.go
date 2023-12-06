package clients

import (
	"accommodation/data"
	"accommodation/domain"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

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

// TODO: Client methods

func (pc ProfileClient) GetUserId(ctx context.Context, username string) (string, error) {
	var timeout time.Duration
	deadline, reqHasDeadline := ctx.Deadline()
	if reqHasDeadline {
		timeout = time.Until(deadline)
	}

	cbResp, err := pc.cb.Execute(func() (interface{}, error) {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, pc.address+"/users/"+username, nil)
		if err != nil {
			return "", err
		}
		return pc.client.Do(req)
	})
	if err != nil {
		return "", handleHttpReqErr(err, pc.address+"/search", http.MethodPost, timeout)
	}

	resp := cbResp.(*http.Response)
	if resp.StatusCode != http.StatusOK {
		return "", domain.ErrResp{
			URL:        resp.Request.URL.String(),
			Method:     resp.Request.Method,
			StatusCode: resp.StatusCode,
		}
	}

	// Parse the JSON response
	var serviceResponse data.User
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&serviceResponse); err != nil {
		return "", fmt.Errorf("failed to decode JSON response: %v", err)
	}

	return serviceResponse.ID.Hex(), nil
}
