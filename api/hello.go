package api

import (
	"fmt"
	"net/http"
)

// HelloServerHandler is a handler that returns a welcome message.
func (s *Server) HelloServerHandler(w http.ResponseWriter, _ *http.Request) {
	fmt.Fprintf(w, "Welcome to my financial server!!!")
}
