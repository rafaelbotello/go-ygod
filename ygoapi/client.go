// Package ygoapi provides a client for interacting with the
// YGOPRODeck API to fetch Yu-Gi-Oh! card data and images.
package ygoapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"math/rand/v2"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/time/rate"
)

const BaseURL = "https://db.ygoprodeck.com/api/v7/cardinfo.php"

var ErrRateLimitExceeded = errors.New("fatal API error (rate limit or forbidden)")

type Client struct {
	baseURL string
	client  *http.Client
	limiter *rate.Limiter
}

func NewClient(baseURL string, client *http.Client) *Client {
	return &Client{
		baseURL: baseURL,
		client:  client,
		limiter: rate.NewLimiter(15, 15),
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

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create cards request: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute cards request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var response GetCardsResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode cards response: %w", err)
	}

	return &response, nil
}

func (c *Client) DownloadImage(ctx context.Context, url string, dest string) error {

	if _, err := os.Stat(dest); err == nil {
		return nil
	}

	const maxRetries = 3
	var resp *http.Response
	var lastErr error

	for attempt := 1; attempt <= maxRetries; attempt++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return fmt.Errorf("failed to create Image request: %w", err)
		}

		err = c.limiter.Wait(ctx)
		if err != nil {
			return fmt.Errorf("rate limiter blocked request: %w", err)
		}

		resp, err = c.client.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("error fetching image: %v", err)
		} else {
			switch resp.StatusCode {
			case http.StatusOK:
				lastErr = nil
			case http.StatusTooManyRequests:
				resp.Body.Close()
				return ErrRateLimitExceeded
			case http.StatusNotFound:
				resp.Body.Close()
				return fmt.Errorf("image not found 404: %s", url)
			default:
				resp.Body.Close()
				lastErr = fmt.Errorf("unexpected status %d", resp.StatusCode)
			}
		}

		if lastErr == nil {
			break
		}

		if attempt < maxRetries {
			delay := time.Duration(math.Pow(2, float64(attempt))*500)*time.Millisecond + time.Duration(rand.IntN(200))*time.Millisecond
			log.Printf("network hiccup for %s, retrying in %v attempt: %d/%d", filepath.Base(url), delay, attempt+1, maxRetries)

			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
			}
		}
	}

	if lastErr != nil {
		return fmt.Errorf("failed after %d attempts : %w", maxRetries, lastErr)
	}
	defer func() {
		if resp != nil && resp.Body != nil {
			resp.Body.Close()
		}
	}()

	out, err := os.Create(dest)
	if err != nil {
		return fmt.Errorf("failed to create file : %w", err)
	}
	defer func() {
		out.Close()
		if err != nil {
			os.Remove(dest)
		}
	}()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to save image data: %w", err)
	}

	return nil
}
