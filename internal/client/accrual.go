package client

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/PrahaTurbo/gophermart/internal/models"
)

type AccrualClient struct {
	baseURL string
	client  *http.Client
}

func NewAccrualClient(baseURL string) *AccrualClient {
	return &AccrualClient{
		baseURL: baseURL,
		client:  &http.Client{},
	}
}

func (c *AccrualClient) GetAccrual(orderID string) (*models.AccrualResponse, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/api/orders/%s", c.baseURL, orderID), nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var accrualResponse models.AccrualResponse
	if err := json.Unmarshal(body, &accrualResponse); err != nil {
		return nil, err
	}

	return &accrualResponse, nil
}
