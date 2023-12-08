package clients

import (
	"auth/data"
	"auth/domain"
	"bytes"
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

// Sends user data to profile service, for persistence in profile_db
// Returns error if it fails
func (c ProfileClient) PassInfoToProfileService(ctx context.Context, info data.NewUser) (interface{}, error) {
	newUser := data.NewUser{
		Username:    info.Username,
		FirstName:   info.FirstName,
		LastName:    info.LastName,
		Email:       info.Email,
		Address:     info.Address,
		Role:        info.Role,
		IsActivated: false,
	}

	requestBody, err := json.Marshal(newUser)
	if err != nil {
		_ = fmt.Errorf("failed to marshal user data: %v", err)
		return nil, err
	}

	var timeout time.Duration
	deadline, reqHasDeadline := ctx.Deadline()
	if reqHasDeadline {
		timeout = time.Until(deadline)
	}

	cbResp, err := c.cb.Execute(func() (interface{}, error) {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.address+"/users", bytes.NewBuffer(requestBody))
		if err != nil {
			return nil, err
		}
		return c.client.Do(req)
	})
	if err != nil {
		return nil, handleHttpReqErr(err, c.address, http.MethodPost, timeout)
	}

	resp := cbResp.(*http.Response)
	if resp.StatusCode != http.StatusCreated {
		return nil, domain.ErrResp{
			URL:        resp.Request.URL.String(),
			Method:     resp.Request.Method,
			StatusCode: resp.StatusCode,
		}
	}

	return true, nil
}
