package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/rafaelbotello/go-ygod/ygoapi"
)

func main() {
	client := ygoapi.NewClient(ygoapi.BaseURL, &http.Client{
		Timeout: 10 * time.Second,
	})

	response, err := client.GetCards()
	if err != nil {
		log.Fatal(err)
	}

	for i, card := range response.Data {
		if i >= 10 {
			break
		}
		fmt.Printf("Card is: %+v imageURL is: %v\n", card.Name, card.CardImages[0].ImageURL)
	}
}
