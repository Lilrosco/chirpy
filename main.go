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
	dbQueries *database.Queries
	appEnv string
}

func main() {
	const filepathRoot = "."
	const port = "8080"
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	appEnv := os.Getenv("PLATFORM")
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
		dbQueries: dbQueries,
		appEnv: appEnv,
	}

	mux.Handle("/app/", apiConfig.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(filepathRoot)))))
	mux.HandleFunc("GET /admin/metrics", apiConfig.handlerMetrics)
	mux.HandleFunc("POST /admin/reset", apiConfig.handlerReset)
	mux.HandleFunc("GET /api/healthz", handlerReadiness)
	mux.HandleFunc("POST /api/validate_chirp", handlerChirpsValidation)
	mux.HandleFunc("POST /api/users", apiConfig.handlerCreateUser)

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
