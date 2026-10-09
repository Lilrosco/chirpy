package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

func main() {
	// Request Multiplexer
	mux := http.NewServeMux()

	// Register catch all / root route
	// mux.HandleFunc("/", rootHandler)

	server := &http.Server{
		Addr: ":8080",
		Handler: mux,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
	}

	fmt.Println("Starting server on port 8080...")

	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

func rootHandler(writer http.ResponseWriter, request *http.Request) {
	if request.URL.Path != "/" {
		http.NotFound(writer, request)
		return
	}

	fmt.Fprint(writer, "Hello, World!")
}
