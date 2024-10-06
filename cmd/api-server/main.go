package main

import (
	"financial-data-web/pkg/handlers"
	"fmt"
	"log"
	"net/http"

	"github.com/joho/godotenv"
)

func main() {
	mux := http.NewServeMux()

	loadEnv()

	mux.HandleFunc("/stocks", handlers.GetStockBySymbolHandler)

	// Start the server
	fmt.Println("server is running on port 8080...")
	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		fmt.Println("error starting server:", err)
	}
}

func loadEnv() {
	err := godotenv.Load("../../.env")
	if err != nil {
		log.Fatal("error loading .env file")
	}
}
