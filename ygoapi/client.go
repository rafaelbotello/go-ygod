// Package ygoapi provides a client for interacting with the
// YGOPRODeck API to fetch Yu-Gi-Oh! card data and images.
package ygoapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

const BaseURL = "https://db.ygoprodeck.com/api/v7/cardinfo.php"

type HTTPClient interface {
	Do(*http.Request) (*http.Response, error)
}

type Client struct {
	BaseURL    string
	httpClient HTTPClient
}

func NewClient(BaseURL string, httpClient HTTPClient) *Client {
	return &Client{
		BaseURL:    BaseURL,
		httpClient: httpClient,
	}
}

type CardImage struct {
	ID       int    `json:"id"`
	ImageURL string `json:"image_url"`
}

type Card struct {
	Name       string      `json:"name"`
	ID         int         `json:"id"`
	CardImages []CardImage `json:"card_images"`
}

type GetCardsResponse struct {
	Data []Card `json:"data"`
}

func (c *Client) GetCards(ctx context.Context) (*GetCardsResponse, error) {

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.BaseURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create Card request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to submit Cards http request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var response *GetCardsResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal Cards http request: %w", err)
	}
	return response, nil
}
