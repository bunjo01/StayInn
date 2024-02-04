package clients

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"reservation/data"
	"reservation/domain"
	"time"

	"github.com/sony/gobreaker"
)

type NotificationClient struct {
	client  *http.Client
	address string
	cb      *gobreaker.CircuitBreaker
}

func NewNotificationClient(client *http.Client, address string, cb *gobreaker.CircuitBreaker) NotificationClient {
	return NotificationClient{
		client:  client,
		address: address,
		cb:      cb,
	}
}

// TODO: Client methods (notify notification service when reservation is created or deleted)

func (nc *NotificationClient) NotifyReservation(ctx context.Context, notification data.Notification, token string) (bool, error) {
	var timeout time.Duration
	deadline, reqHasDeadline := ctx.Deadline()
	if reqHasDeadline {
		timeout = time.Until(deadline)
	}

	requestBody, err := json.Marshal(notification)
	if err != nil {
		return false, fmt.Errorf("failed to marshal notification: %v", err)
	}

	cbResp, err := nc.cb.Execute(func() (interface{}, error) {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, nc.address+"/reservation", bytes.NewBuffer(requestBody))
		if err != nil {
			return false, err
		}
		req.Header.Set("Authorization", "Bearer "+token)
		return nc.client.Do(req)
	})
	if err != nil {
		return false, handleHttpReqErr(err, nc.address+"/reservation", http.MethodPost, timeout)
	}

	resp := cbResp.(*http.Response)
	if resp.StatusCode != http.StatusCreated {
		return false, domain.ErrResp{
			URL:        resp.Request.URL.String(),
			Method:     resp.Request.Method,
			StatusCode: resp.StatusCode,
		}
	}

	return true, nil
}
