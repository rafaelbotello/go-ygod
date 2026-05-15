package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/rafaelbotello/go-ygod/ygoapi"
)

func main() {

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	client := ygoapi.NewClient(ygoapi.BaseURL, http.DefaultClient)
	log.Println("Fetching card data from YGOAPI...")

	response, err := client.GetCards(ctx)
	if err != nil {
		log.Fatalf("Fatal error fetching cards: %v", err)
	}

	err = os.MkdirAll("images/", 0755)
	if err != nil {
		log.Fatalf("Failed to create directory: %v", err)
	}

	var urls []string

	for _, card := range response.Data {
		urls = append(urls, card.CardImages[0].ImageURL)
	}

	log.Printf("Starting download of %d images...", len(urls))

	err = client.DownloadAllImages(ctx, urls, "images/", 4)
	if err != nil {
		if errors.Is(err, ygoapi.ErrRateLimitExceeded) {
			log.Fatalf("Factory shut down early due to API Rate Limiting: %v", err)
		}
		log.Fatalf("Factory shut down with error: %v", err)
	}

	log.Println("All downloads complete! The factory is closed.")

}
