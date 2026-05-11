package ygoapi_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/rafaelbotello/go-ygod/ygoapi"
	mock_ygoapi "github.com/rafaelbotello/go-ygod/ygoapi/mock_httpclient"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestGetCards(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockHTTPClient := mock_ygoapi.NewMockHTTPClient(ctrl)

	testBaseURL := "http://mybaseurl.com"

	_, err := http.NewRequest(http.MethodGet, testBaseURL, nil)
	require.NoError(t, err)

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

	jsonBytes, err := json.Marshal(expectedResponse)
	require.NoError(t, err)

	client := ygoapi.NewClient(testBaseURL, mockHTTPClient)

	t.Run("Success Download", func(t *testing.T) {

		mockHTTPClient.EXPECT().Do(gomock.Any()).Return(&http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewBuffer(jsonBytes)),
		}, nil)

		response, err := client.GetCards(t.Context())
		require.NoError(t, err)
		require.Equal(t, expectedResponse, response)

	})

	t.Run("Fails on 429", func(t *testing.T) {

		mockHTTPClient.EXPECT().Do(gomock.Any()).Return(&http.Response{
			StatusCode: http.StatusTooManyRequests,
			Body:       io.NopCloser(bytes.NewBuffer(jsonBytes)),
		}, nil)

		_, err := client.GetCards(t.Context())
		require.Error(t, err)
		require.Contains(t, err.Error(), "429")
	})

	t.Run("Fails on 500", func(t *testing.T) {

		mockHTTPClient.EXPECT().Do(gomock.Any()).Return(&http.Response{
			StatusCode: http.StatusInternalServerError,
			Body:       io.NopCloser(bytes.NewBufferString("Server crashed")),
		}, nil)

		_, err := client.GetCards(t.Context())
		require.Error(t, err)
		require.Contains(t, err.Error(), "500")
	})

	t.Run("Fails on Timeout", func(t *testing.T) {

		timeoutErr := errors.New("network timeout")

		mockHTTPClient.EXPECT().Do(gomock.Any()).Return(nil, timeoutErr)

		_, err := client.GetCards(t.Context())

		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to submit Cards http request")
		require.Contains(t, err.Error(), "network timeout")
	})

}

func TestImageDownload(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockHTTPClient := mock_ygoapi.NewMockHTTPClient(ctrl)

	testBaseURL := "http://mybaseurl.com"

	client := ygoapi.NewClient(testBaseURL, mockHTTPClient)

	t.Run("Download Success", func(t *testing.T) {

		fakeImageBytes := []byte("fake-image-data-12345")

		mockHTTPClient.EXPECT().Do(gomock.Any()).Return(&http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewBuffer(fakeImageBytes)),
		}, nil)

		tempDir := t.TempDir()
		destPath := filepath.Join(tempDir, "123456_test.jpg")

		err := client.DownloadImage(t.Context(), "http://mybaseurl.com/image.jpg", destPath)
		require.NoError(t, err)

		savedBytes, err := os.ReadFile(destPath)
		require.NoError(t, err, "failed to read the save file")
		require.Equal(t, fakeImageBytes, savedBytes, "saved file bytes do not match mock bytes")

	})

	t.Run("Fails on 404", func(t *testing.T) {
		mockHTTPClient.EXPECT().Do(gomock.Any()).Return(&http.Response{
			StatusCode: http.StatusNotFound,
			Body:       io.NopCloser(bytes.NewBufferString("Not Found")),
		}, nil)

		tempDir := t.TempDir()
		destPath := filepath.Join(tempDir, "fail_card.jpg")

		err := client.DownloadImage(t.Context(), "http://mybaseurl.com/image.jpg", destPath)
		require.Error(t, err)
		require.Contains(t, err.Error(), "404")

		_, fileErr := os.Stat(destPath)
		require.True(t, os.IsNotExist(fileErr), "file should not exist on a 404 error")
	})

	t.Run("Fails on 429", func(t *testing.T) {
		mockHTTPClient.EXPECT().Do(gomock.Any()).Return(&http.Response{
			StatusCode: http.StatusTooManyRequests,
			Body:       io.NopCloser(bytes.NewBufferString("Not Found")),
		}, nil)

		tempDir := t.TempDir()
		destPath := filepath.Join(tempDir, "fail_card.jpg")

		err := client.DownloadImage(t.Context(), "http://mybaseurl.com/image.jpg", destPath)
		require.Error(t, err)
		require.Contains(t, err.Error(), "429")

		_, fileErr := os.Stat(destPath)
		require.True(t, os.IsNotExist(fileErr), "file should not exist on a 429 error")
	})

	t.Run("Fails on 403", func(t *testing.T) {
		mockHTTPClient.EXPECT().Do(gomock.Any()).Return(&http.Response{
			StatusCode: http.StatusForbidden,
			Body:       io.NopCloser(bytes.NewBufferString("Not Found")),
		}, nil)

		tempDir := t.TempDir()
		destPath := filepath.Join(tempDir, "fail_card.jpg")

		err := client.DownloadImage(t.Context(), "http://mybaseurl.com/image.jpg", destPath)
		require.Error(t, err)
		require.Contains(t, err.Error(), "403")

		_, fileErr := os.Stat(destPath)
		require.True(t, os.IsNotExist(fileErr), "file should not exist on a 429 error")
	})

}
