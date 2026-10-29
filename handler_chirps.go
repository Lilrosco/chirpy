package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"sort"
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
	}

	userID, err := cfg.getUserIDFromBearerToken(w, r)

	if err != nil {
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
			Body:   cleanedBody,
			UserID: userID,
		},
	)

	if err != nil {
		log.Printf("Could not create chirp: %v\n", err)
		respondWithError(w, http.StatusInternalServerError, "Internal Server Error", err)
		return
	}

	respondWithJSON(w, http.StatusCreated, Chirp{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserID:    chirp.UserID,
	})
}

func (cfg *apiConfig) handlerGetChirps(w http.ResponseWriter, r *http.Request) {
	userID, err := getAuthorIDFromRequest(r)

	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid Author ID", err)
		return
	}

	ordering := getSortingFromRequest(r)

	var dbChirps []database.Chirp

	if userID != uuid.Nil {
		dbChirps, err = cfg.db.GetChirpsForUser(r.Context(), userID)
	} else {
		dbChirps, err = cfg.db.GetChirps(r.Context())
	}

	if ordering == "asc" {
		sort.Slice(dbChirps, func(i, j int) bool { return dbChirps[i].CreatedAt.Before(dbChirps[j].CreatedAt) })
	} else {
		sort.Slice(dbChirps, func(i, j int) bool { return dbChirps[i].CreatedAt.After(dbChirps[j].CreatedAt) })
	}

	if err != nil {
		log.Printf("Could not fetch all chirps: %v\n", err)
		respondWithError(w, http.StatusInternalServerError, "Internal Server Error", err)
		return
	}

	chirps := []Chirp{}

	for _, chirp := range dbChirps {
		chirps = append(chirps, Chirp{
			ID:        chirp.ID,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
			Body:      chirp.Body,
			UserID:    chirp.UserID,
		})
	}

	respondWithJSON(w, http.StatusOK, chirps)
}

func (cfg *apiConfig) handlerGetSingleChirp(w http.ResponseWriter, r *http.Request) {
	chirpIDStr := r.PathValue("chirpID")
	chirpID, err := uuid.Parse(chirpIDStr)

	if err != nil {
		log.Printf("Could not parse passed in UUID: %s | %v\n", chirpIDStr, err)
		respondWithError(w, http.StatusInternalServerError, "Internal Server Error", err)
		return
	}

	dbChirp, err := cfg.db.GetChirpByID(r.Context(), chirpID)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondWithError(w, http.StatusNotFound, "Chirp with specified ID not found", err)
			return
		}

		log.Printf("Could not fetch chirp with ID: %s | %v\n", chirpIDStr, err)
		respondWithError(w, http.StatusInternalServerError, "Internal Server Error", err)
		return
	}

	respondWithJSON(w, http.StatusOK, Chirp{
		ID:        dbChirp.ID,
		CreatedAt: dbChirp.CreatedAt,
		UpdatedAt: dbChirp.UpdatedAt,
		Body:      dbChirp.Body,
		UserID:    dbChirp.UserID,
	})
}

func (cfg *apiConfig) handlerDeleteChirp(w http.ResponseWriter, r *http.Request) {
	userID, err := cfg.getUserIDFromBearerToken(w, r)

	if err != nil {
		return
	}

	chirpIDStr := r.PathValue("chirpID")
	chirpID, err := uuid.Parse(chirpIDStr)

	if err != nil {
		log.Printf("Could not parse passed in UUID: %s | %v\n", chirpIDStr, err)
		respondWithError(w, http.StatusBadRequest, "Invalid Chirp ID", err)
		return
	}

	chirp, err := cfg.db.GetChirpByID(r.Context(), chirpID)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondWithError(w, http.StatusNotFound, "Chirp with specified ID not found", err)
			return
		}

		log.Printf("Could not fetch chirp with ID: %s | %v\n", chirpIDStr, err)
		respondWithError(w, http.StatusInternalServerError, "Internal Server Error", err)
		return
	} else if chirp.UserID != userID {
		respondWithError(w, http.StatusForbidden, "Forbidden", errors.New("User if not the author of chirp"))
		return
	}

	err = cfg.db.DeleteChirp(r.Context(), chirpID)

	if err != nil {
		log.Printf("Could not delete chirp with ID: %s | %v\n", chirpIDStr, err)
		respondWithError(w, http.StatusInternalServerError, "Internal Server Error", err)
		return
	}

	respondWithJSON(w, http.StatusNoContent, nil)
}

func getAuthorIDFromRequest(r *http.Request) (uuid.UUID, error) {
	userIDStr := r.URL.Query().Get("author_id")

	if userIDStr == "" {
		return uuid.Nil, nil
	}

	userID, err := uuid.Parse(userIDStr)

	if err != nil {
		return uuid.Nil, err
	}

	return userID, nil
}

func getSortingFromRequest(r *http.Request) string {
	sort := r.URL.Query().Get("sort")

	if sort == "" {
		return "asc"
	}

	sort = strings.ToLower(sort)

	if sort != "asc" && sort != "desc" {
		return "asc"
	}

	return sort
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
	badWords := map[string]struct{}{
		"kerfuffle": {},
		"sharbert":  {},
		"fornax":    {},
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
