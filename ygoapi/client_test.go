package ygoapi_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/rafaelbotello/go-ygod/ygoapi"
	mock_ygoapi "github.com/rafaelbotello/go-ygod/ygoapi/mock_httpclient"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestClient(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockHttpClient := mock_ygoapi.NewMockHttpClient(ctrl)

	testBaseURL := "http://mybaseurl.com"
	expectedRequest, err := http.NewRequest(http.MethodGet, testBaseURL, nil)
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

	mockHttpClient.EXPECT().Do(expectedRequest).Return(&http.Response{
		Body: io.NopCloser(bytes.NewBuffer(jsonBytes)),
	}, nil)

	client := ygoapi.NewClient(testBaseURL, mockHttpClient)
	response, err := client.GetCards()
	require.NoError(t, err)
	require.Equal(t, expectedResponse, response)

}
