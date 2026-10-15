package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

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
		UserID uuid.UUID `json:"user_id"`
    }

	type response struct {
		Chirp
	}

	decoder := json.NewDecoder(r.Body)
    params := parameters{}
    err := decoder.Decode(&params)

    if err != nil {
		respondWithError(w, http.StatusInternalServerError, "couldn't decode parameters", err)
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
			UserID: params.UserID,
		},
	)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "could not create chirp", err)
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
		respondWithError(w, http.StatusInternalServerError, "could not fetch all chirps", err)
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
