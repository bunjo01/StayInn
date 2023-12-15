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
	"go.mongodb.org/mongo-driver/bson/primitive"
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

// TODO: Client methods (checking username and ID)

func (pc ProfileClient) GetUserId(ctx context.Context, username, token string) (string, error) {
	var timeout time.Duration
	deadline, reqHasDeadline := ctx.Deadline()
	if reqHasDeadline {
		timeout = time.Until(deadline)
	}

	cbResp, err := pc.cb.Execute(func() (interface{}, error) {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, pc.address+"/users/"+username, nil)
		if err != nil {
			return "", err
		}
		req.Header.Set("Authorization", "Bearer "+token)
		return pc.client.Do(req)
	})
	if err != nil {
		return "", handleHttpReqErr(err, pc.address+"/users/"+username, http.MethodPost, timeout)
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

func (pc ProfileClient) GetUserById(ctx context.Context, id primitive.ObjectID, token string) (data.User, error) {
	var timeout time.Duration
	deadline, reqHasDeadline := ctx.Deadline()
	if reqHasDeadline {
		timeout = time.Until(deadline)
	}

	idUser := data.UserId{}
	idUser.ID = id
	requestBody, err := json.Marshal(idUser)
	if err != nil {
		return data.User{}, fmt.Errorf("failed to marshal dates: %v", err)
	}

	cbResp, err := pc.cb.Execute(func() (interface{}, error) {
		fmt.Println("Profile service address:", pc.address)
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, pc.address+"/users/get-user-by-id", bytes.NewBuffer(requestBody))
		if err != nil {
			return "", err
		}
		req.Header.Set("Authorization", "Bearer "+token)
		return pc.client.Do(req)
	})
	if err != nil {
		return data.User{}, handleHttpReqErr(err, pc.address+"/users/get-username-by-id", http.MethodPost, timeout)
	}

	resp := cbResp.(*http.Response)
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Error: Profile service returned status code %d for fetching user by id", resp.StatusCode)
		return data.User{}, domain.ErrResp{
			URL:        resp.Request.URL.String(),
			Method:     resp.Request.Method,
			StatusCode: resp.StatusCode,
		}
	}

	// Parse the JSON response
	var serviceResponse data.User
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&serviceResponse); err != nil {
		return data.User{}, fmt.Errorf("failed to decode JSON response: %v", err)
	}

	return serviceResponse, nil
}
