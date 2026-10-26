package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/Lilrosco/chirpy/internal/auth"
	"github.com/Lilrosco/chirpy/internal/database"
	"github.com/google/uuid"
)

type Chirp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
}

func (cfg *apiConfig) handlerCreateChirp(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body string `json:"body"`
	}

	type response struct {
		Chirp
	}

	token, err := auth.GetBearerToken(r.Header)

	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Missing authorization", err)
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)

	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized!", err)
		return
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err = decoder.Decode(&params)

	if err != nil {
		log.Printf("Could not decode parameters: %v\n", err)
		respondWithError(w, http.StatusInternalServerError, "Internal Server Error", err)
		return
	}

	cleanedBody, err := validateChirp(params.Body)

	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	chirp, err := cfg.db.CreateChirp(
		r.Context(),
		database.CreateChirpParams{
			Body: cleanedBody,
			UserID: userID,
		},
	)

	if err != nil {
		log.Printf("Could not create chirp: %v\n", err)
		respondWithError(w, http.StatusInternalServerError, "Internal Server Error", err)
		return
	}

	respondWithJSON(w, http.StatusCreated, response{
		Chirp: Chirp{
			ID: chirp.ID,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
			Body: chirp.Body,
			UserID: chirp.UserID,
		},
	})
}

func (cfg *apiConfig) handlerGetChirps(w http.ResponseWriter, r *http.Request) {
	dbChirps, err := cfg.db.GetChirps(r.Context())

	if err != nil {
		log.Printf("Could not fetch all chirps: %v\n", err)
		respondWithError(w, http.StatusInternalServerError, "Internal Server Error", err)
		return
	}

	chirps := []Chirp{}

	for _, chirp := range dbChirps {
		chirps = append(chirps, Chirp{
			ID: chirp.ID,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
			Body: chirp.Body,
			UserID: chirp.UserID,
		})
	}

	respondWithJSON(w, http.StatusOK, chirps)
}

func (cfg *apiConfig) handlerGetSingleChirp(w http.ResponseWriter, r *http.Request) {
	type response struct {
		Chirp
	}

	chirpID := r.PathValue("chirpID")
	id, err := uuid.Parse(chirpID)

	if err != nil {
		log.Printf("Could not parse passed in UUID: %s | %v\n", chirpID, err)
		respondWithError(w, http.StatusInternalServerError, "Internal Server Error", err)
	}

	dbChirp, err := cfg.db.GetChirpByID(r.Context(), id)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondWithError(w, http.StatusNotFound, "Chirp with specified ID not found", err)
			return
		}

		log.Printf("Could not fetch chirp with ID: %s | %v\n", chirpID, err)
		respondWithError(w, http.StatusInternalServerError, "Internal Server Error", err)
		return
	}

	respondWithJSON(w, http.StatusOK, response{
		Chirp: Chirp{
			ID: dbChirp.ID,
			CreatedAt: dbChirp.CreatedAt,
			UpdatedAt: dbChirp.UpdatedAt,
			Body: dbChirp.Body,
			UserID: dbChirp.UserID,
		},
	})
}

func validateChirp(body string) (string, error) {
	maxChirpLen := 140
	chirpLen := len(body)

	if chirpLen > maxChirpLen {
		return "", errors.New("Chirp too long")
	} else if chirpLen == 0 {
		return "", errors.New("Chirp cannot be empty")
	}

	cleanedBody := cleanBody(body)
	return cleanedBody, nil
}

func cleanBody(body string) string {
	badWords := map[string]struct{} {
		"kerfuffle" : {},
		"sharbert" : {},
		"fornax" : {},
	}
	words := strings.Split(body, " ")

	for index, word := range words {
		_, exists := badWords[strings.ToLower(word)]

		if exists {
			words[index] = "****"
		}
	}

	return strings.Join(words, " ")
}
