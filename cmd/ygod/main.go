package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"

	"github.com/rafaelbotello/go-ygod/ygoapi"
	"github.com/schollz/progressbar/v3"
)

func main() {

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
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

	errorFile, err := os.OpenFile("failed_images.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Failed to open error log file: %v", err)
	}
	defer errorFile.Close()

	errorLogger := log.New(errorFile, "", log.Ldate|log.Ltime)

	log.Printf("Starting download of %d images...", len(urls))
	bar := progressbar.Default(int64(len(urls)), "Downloading Cards")

	err = client.DownloadAllImages(ctx, urls, "images/", 20, bar, errorLogger)
	if err != nil {
		if errors.Is(err, ygoapi.ErrRateLimitExceeded) {
			log.Fatalf("Factory shut down early due to API Rate Limiting: %v", err)
		}
		log.Fatalf("Factory shut down with error: %v", err)
	}

	log.Println("All downloads complete! The factory is closed.")

}
