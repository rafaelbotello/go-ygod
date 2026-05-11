package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/rafaelbotello/go-ygod/ygoapi"
)

func main() {

	ctx := context.Background()

	client := ygoapi.NewClient(ygoapi.BaseURL, &http.Client{
		Timeout: 10 * time.Second,
	})

	response, err := client.GetCards(ctx)
	if err != nil {
		log.Fatal(err)
	}

	err = os.MkdirAll("images/", 0755)
	if err != nil {
		log.Fatalf("failed to create directory: %w", err)
	}

	for i, card := range response.Data {
		if i >= 10 {
			break
		}

		imageURL := card.CardImages[0].ImageURL
		fileName := fmt.Sprintf("%d.jpg", card.ID)
		fullPath := filepath.Join("images", fileName)

		err := client.DownloadImage(ctx, imageURL, fullPath)

		if err != nil {
			fmt.Printf("error: %v\n", err)
		}

		fmt.Printf("%v Download Complete!\n", card.ID)

	}
}
