package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/huy125/financial-data-web/api"
	repository "github.com/huy125/financial-data-web/api/repositories/in-memory"
)

func main() {
	var apiKey string
	flag.StringVar(&apiKey, "apiKey", "", "Alpha Vantage API Key, required for stocks endpoints")
	flag.Parse()

	if apiKey == "" {
		log.Fatal("apiKey is required")
	}

	userRepo := repository.NewInMemoryUserRepository()
	srv := api.New(apiKey, userRepo)

	log.Println("Starting server on port :8080")
	err := http.ListenAndServe(":8080", srv)
	if err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
