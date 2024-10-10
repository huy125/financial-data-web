package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/huy125/financial-data-web/api"
)

func main() {
	var apiKey string
	flag.StringVar(&apiKey, "apiKey", "", "Alpha Vantage API Key, required for stocks endpoints")
	flag.Parse()

	if apiKey == "" {
		log.Fatal("apiKey is required")
	}

	srv := api.New(apiKey)

	log.Println("Starting server on port :8080")
	err := http.ListenAndServe(":8080", srv)
	if err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
