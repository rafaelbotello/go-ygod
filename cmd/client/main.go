package main

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"time"

	"path/filepath"

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

	for i, card := range response.Data {
		if i >= 10 {
			break
		}

		imageURL := card.CardImages[0].ImageURL
		fileName := fmt.Sprintf("%d.jpg", card.ID)
		fullPath := filepath.Join("images", fileName)

		err := downloadImage(imageURL, fullPath)
		if err != nil {
			fmt.Printf("error: %v\n", err)
		}

		fmt.Printf("%v Download Complete!\n", card.ID)

	}
}

func downloadImage(url string, dest string) error {

	err := os.MkdirAll("images/", 0755)
	if err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("error fetching image: %v\n", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	const mode fs.FileMode = 0644

	out, err := os.OpenFile(dest, os.O_CREATE|os.O_WRONLY, mode)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}
