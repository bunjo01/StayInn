package clients

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"profile/domain"
	"time"

	"github.com/sony/gobreaker"
)

type AuthClient struct {
	client  *http.Client
	address string
	cb      *gobreaker.CircuitBreaker
}

func NewAuthClient(client *http.Client, address string, cb *gobreaker.CircuitBreaker) AuthClient {
	return AuthClient{
		client:  client,
		address: address,
		cb:      cb,
	}
}

// TODO: Client methods

// Send changed username to auth service
// Returns error if it fails
func (c AuthClient) PassUsernameToAuthService(ctx context.Context, oldUsername, username string) (interface{}, error) {
	reqBody := map[string]string{"username": username}
	requestBody, err := json.Marshal(reqBody)
	if err != nil {
		_ = fmt.Errorf("error marshaling request body: %s", err)
		return nil, err
	}

	var timeout time.Duration
	deadline, reqHasDeadline := ctx.Deadline()
	if reqHasDeadline {
		timeout = time.Until(deadline)
	}

	cbResp, err := c.cb.Execute(func() (interface{}, error) {
		req, err := http.NewRequestWithContext(ctx, http.MethodPut,
			c.address+"/update-username"+"/"+oldUsername+"/"+username, bytes.NewBuffer(requestBody))
		if err != nil {
			return nil, err
		}
		return c.client.Do(req)
	})
	if err != nil {
		return nil, handleHttpReqErr(err, c.address, http.MethodPut, timeout)
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
