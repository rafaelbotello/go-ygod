package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"

	"github.com/rafaelbotello/go-ygod/ygoapi"
	"github.com/schollz/progressbar/v3"
)

func main() {

	textLogger := slog.NewTextHandler(os.Stdout, nil)
	slog.SetDefault(slog.New(textLogger))

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	client := ygoapi.NewClient(ygoapi.BaseURL, http.DefaultClient, slog.Default())
	slog.Info("Fetching card data from YGOAPI...")

	response, err := client.GetCards(ctx)
	if err != nil {
		slog.Error("Fatal error fetching cards", "error", err)
		os.Exit(1)
	}

	err = os.MkdirAll("images/", 0755)
	if err != nil {
		slog.Error("Failed to create directory", "error", err)
		os.Exit(1)
	}

	var urls []string

	for _, card := range response.Data {
		urls = append(urls, card.CardImages[0].ImageURL)
	}

	errorFile, err := os.OpenFile("failed_images.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		slog.Error("Failed to open error log file", "error", err)
		os.Exit(1)
	}
	defer errorFile.Close()

	jsonFileHandler := slog.NewJSONHandler(errorFile, nil)
	errorLogger := slog.New(jsonFileHandler)

	slog.Info("Starting download", "total_images", len(urls))
	bar := progressbar.Default(int64(len(urls)), "Downloading Cards")

	err = client.DownloadAllImages(ctx, urls, "images/", 20, bar, errorLogger)
	if err != nil {
		errorFile.Close()
		if errors.Is(err, ygoapi.ErrRateLimitExceeded) {
			slog.Error("Factory shut down early due to API Rate Limiting", "error", err)
			os.Exit(1)
		}
		slog.Error("Factory shut down with error", "error", err)
		os.Exit(1)
	}

	slog.Info("All downloads complete! The factory is closed.")

}
