package main

import (
	"financial-data-web/pkg/handlers"
	"fmt"
	"net/http"
)



func main() {
	mux := http.NewServeMux()

	// routes
	mux.HandleFunc("/", handlers.GetStockHandler)

	// Start the server
	fmt.Println("Server is running on port 8080...")
	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		fmt.Println("Error starting server:", err)
	}
}