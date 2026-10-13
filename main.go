package main

import (
	"fmt"
	"log"
	"net/http"
	"sync/atomic"
)

type apiConfig struct {
	fileserverHits atomic.Int32
}

func main() {
	const filepathRoot = "."
	const port = "8080"

	// Request Multiplexer
	mux := http.NewServeMux()

	// Register catch all / root route
	// mux.HandleFunc("/", rootHandler)

	apiConfig := &apiConfig{
		fileserverHits: atomic.Int32{},
	}

	mux.Handle("/app/", apiConfig.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(filepathRoot)))))
	mux.HandleFunc("GET /admin/metrics", apiConfig.handlerMetrics)
	mux.HandleFunc("POST /admin/reset", apiConfig.handlerReset)
	mux.HandleFunc("GET /api/healthz", handlerReadiness)
	mux.HandleFunc("POST /api/validate_chirp", handlerChirpsValidation)

	server := &http.Server{
		Addr: ":" + port,
		Handler: mux,
	}

	fmt.Println("Starting server on port 8080...")

	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

func handlerRoot(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	fmt.Fprint(w, "Hello, World!")
}
