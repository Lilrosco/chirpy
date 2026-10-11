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

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) handlerMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("Hits: %d", cfg.fileserverHits.Load())))
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
	mux.HandleFunc("GET /api/metrics", apiConfig.handlerMetrics)
	mux.HandleFunc("POST /api/reset", apiConfig.handlerReset)
	mux.HandleFunc("GET /api/healthz", handlerReadiness)

	server := &http.Server{
		Addr: ":" + port,
		Handler: mux,
	}

	fmt.Println("Starting server on port 8080...")

	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

func handlerRoot(writer http.ResponseWriter, request *http.Request) {
	if request.URL.Path != "/" {
		http.NotFound(writer, request)
		return
	}

	fmt.Fprint(writer, "Hello, World!")
}
