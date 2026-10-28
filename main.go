package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/Lilrosco/chirpy/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

// For pq database error
const errUniqueViolation = "23505"

type apiConfig struct {
	fileserverHits atomic.Int32
	db             *database.Queries
	platform       string
	jwtSecret      string
}

func main() {
	const filepathRoot = "."
	const port = "8080"

	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	platform := os.Getenv("PLATFORM")
	jwtSecret := os.Getenv("JWT_SECRET")

	if dbURL == "" {
		log.Fatal("DB_URL must be set")
	}

	if platform == "" {
		log.Fatal("PLATFORM must be set")
	}

	if jwtSecret == "" {
		log.Fatal("JWT_SECRET must be set")
	}

	db, err := sql.Open("postgres", dbURL)

	if err != nil {
		log.Fatalf("Server failed to start DB connection: %v", err)
	}

	dbQueries := database.New(db)

	// Request Multiplexer
	mux := http.NewServeMux()

	// Register catch all / root route
	// mux.HandleFunc("/", rootHandler)

	apiConfig := &apiConfig{
		fileserverHits: atomic.Int32{},
		db:             dbQueries,
		platform:       platform,
		jwtSecret:      jwtSecret,
	}

	mux.Handle("/app/", apiConfig.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(filepathRoot)))))
	mux.HandleFunc("POST /api/chirps", apiConfig.handlerCreateChirp)
	mux.HandleFunc("POST /api/users", apiConfig.handlerCreateUser)
	mux.HandleFunc("PUT /api/users", apiConfig.handlerUpdateUser)
	mux.HandleFunc("POST /api/login", apiConfig.handlerLogin)
	mux.HandleFunc("POST /api/refresh", apiConfig.handlerRefreshToken)
	mux.HandleFunc("POST /api/revoke", apiConfig.handlerRevokeToken)
	mux.HandleFunc("GET /api/chirps", apiConfig.handlerGetChirps)
	mux.HandleFunc("GET /api/chirps/{chirpID}", apiConfig.handlerGetSingleChirp)
	mux.HandleFunc("DELETE /api/chirps/{chirpID}", apiConfig.handlerDeleteChirp)
	mux.HandleFunc("GET /api/healthz", handlerReadiness)
	mux.HandleFunc("POST /api/polka/webhooks", apiConfig.handlerPolkaWebhook)
	mux.HandleFunc("GET /admin/metrics", apiConfig.handlerMetrics)
	mux.HandleFunc("POST /admin/reset", apiConfig.handlerReset)

	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Println("Starting server on port 8080...")

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
