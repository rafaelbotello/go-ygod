// Package ygoapi provides a client for interacting with the
// YGOPRODeck API to fetch Yu-Gi-Oh! card data and images.
package ygoapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
)

const BaseURL = "https://db.ygoprodeck.com/api/v7/cardinfo.php"

var ErrFatalAPI = errors.New("fatal API error (rate limit or forbidden)")

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

func (c *Client) DownloadImage(ctx context.Context, url string, dest string) error {

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("failed to create Image request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("error fetching image: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode == http.StatusForbidden {
		return fmt.Errorf("%w: received status %d", ErrFatalAPI, resp.StatusCode)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	out, err := os.Create(dest)
	if err != nil {
		return fmt.Errorf("failed to save Image: %w", err)
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}
