package ygoapi_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"github.com/rafaelbotello/go-ygod/ygoapi"
	"github.com/stretchr/testify/require"
)

func TestGetCards(t *testing.T) {

	t.Run("success", func(t *testing.T) {
		expectedResponse := &ygoapi.GetCardsResponse{
			Data: []ygoapi.Card{
				{
					ID:   80181649,
					Name: "A Case for K9",
					CardImages: []ygoapi.CardImage{
						{
							ID:       34541863,
							ImageURL: "https://images.ygoprodeck.com/images/cards/80181649.jpg",
						},
					},
				},
			},
		}

		server := httptest.NewServer(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, http.MethodGet, r.Method)
				require.Equal(t, "/", r.URL.Path)
				w.WriteHeader(http.StatusOK)
				err := json.NewEncoder(w).Encode(expectedResponse)
				require.NoError(t, err)
			}),
		)
		defer server.Close()

		client := ygoapi.NewClient(server.URL, server.Client())

		response, err := client.GetCards(t.Context())

		require.NoError(t, err)
		require.Equal(t, expectedResponse, response)
	})

	t.Run("fails on 429", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusTooManyRequests)
		}),
		)
		defer server.Close()

		client := ygoapi.NewClient(server.URL, server.Client())

		_, err := client.GetCards(t.Context())

		require.Error(t, err)
		require.Contains(t, err.Error(), "429")
	})

	t.Run("fails on 500", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}),
		)
		defer server.Close()

		client := ygoapi.NewClient(server.URL, server.Client())

		_, err := client.GetCards(t.Context())

		require.Error(t, err)
		require.Contains(t, err.Error(), "500")
	})

	t.Run("fails on invalid json", func(t *testing.T) {

		server := httptest.NewServer(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("{invalid-json"))
			}),
		)
		defer server.Close()

		client := ygoapi.NewClient(server.URL, server.Client())

		_, err := client.GetCards(t.Context())

		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to decode")
	})
}

func TestDownloadImage(t *testing.T) {

	t.Run("download success", func(t *testing.T) {
		fakeImageBytes := []byte("fake-image-data-12345")
		server := httptest.NewServer(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, http.MethodGet, r.Method)
				w.WriteHeader(http.StatusOK)

				_, err := w.Write(fakeImageBytes)
				require.NoError(t, err)
			}),
		)
		defer server.Close()

		client := ygoapi.NewClient(server.URL, server.Client())

		destPath := filepath.Join(
			t.TempDir(),
			"test_card.jpg",
		)

		err := client.DownloadImage(
			t.Context(),
			server.URL+"/image.jpg",
			destPath,
		)

		require.NoError(t, err)

		savedBytes, err := os.ReadFile(destPath)

		require.NoError(t, err)
		require.Equal(t, fakeImageBytes, savedBytes)
	})

	t.Run("fails on 404", func(t *testing.T) {
		server := httptest.NewServer(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNotFound)
			}),
		)
		defer server.Close()

		client := ygoapi.NewClient(server.URL, server.Client())

		destPath := filepath.Join(t.TempDir(), "fail_card.jpg")

		err := client.DownloadImage(t.Context(), server.URL+"/missing.jpg", destPath)

		require.Error(t, err)
		require.Contains(t, err.Error(), "404")

		_, statErr := os.Stat(destPath)

		require.True(t, os.IsNotExist(statErr))
	})

	t.Run("fails with fatal api error on 429", func(t *testing.T) {

		server := httptest.NewServer(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusTooManyRequests)
			}),
		)
		defer server.Close()

		client := ygoapi.NewClient(server.URL, server.Client())

		destPath := filepath.Join(t.TempDir(), "fail_card.jpg")

		err := client.DownloadImage(t.Context(), server.URL+"/rate_limited.jpg", destPath)

		require.Error(t, err)
		require.ErrorIs(t, err, ygoapi.ErrRateLimitExceeded)
		require.Contains(t, err.Error(), "429")
	})
}

func TestDownloadAllImages(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte("fake_image_data"))
		require.NoError(t, err)
	}))
	defer mockServer.Close()

	client := ygoapi.NewClient(mockServer.URL, mockServer.Client())
	destDir := t.TempDir()

	urls := []string{
		mockServer.URL + "/card1.jpg",
		mockServer.URL + "/card2.jpg",
		mockServer.URL + "/card3.jpg",
		mockServer.URL + "/card4.jpg",
	}

	client.DownloadAllImages(t.Context(), urls, destDir, 15)

	files, err := os.ReadDir(destDir)
	if err != nil {
		t.Fatalf("failed to read temp directory: %v", err)
	}

	if len(files) != 4 {
		t.Errorf("expected 4 files downloaded, got %d", len(files))
	}

}

func TestDownloadAllImages_RateLimitCancellation(t *testing.T) {

	var requestCount int32
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&requestCount, 1)
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer mockServer.Close()

	client := ygoapi.NewClient(mockServer.URL, mockServer.Client())
	destDir := t.TempDir()

	var urls []string

	for range 50 {
		urls = append(urls, mockServer.URL+"/card.jpg")
	}

	client.DownloadAllImages(t.Context(), urls, destDir, 4)

	finalCount := atomic.LoadInt32(&requestCount)

	if finalCount == 50 {
		t.Fatalf("Cancellation failed! The server was hit for all 50 URLs.")
	}

	if finalCount > 10 {
		t.Fatalf("Cancellation took too long! Server was hit %d times", finalCount)
	}

	t.Logf("Success! The panic button worked. Workers stopped after only %d requests instead of 50.", finalCount)
}

func TestDownloadAllImages_RateLimiterPacing(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := ygoapi.NewClient(server.URL, server.Client())
	destDir := t.TempDir()

	var urls []string
	for range 45 {
		urls = append(urls, server.URL+"/test.jpg")
	}

	startTime := time.Now()

	err := client.DownloadAllImages(t.Context(), urls, destDir, 20)
	require.NoError(t, err)

	elapsed := time.Since(startTime)

	if elapsed < 1500*time.Millisecond {
		t.Fatalf("Rate limiter failed downloaded 45 images way too fast: %v", elapsed)
	}

	t.Logf("Success: rate limiter properly paced 45 requests over %v", elapsed)
}
