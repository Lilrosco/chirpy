package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/Lilrosco/chirpy/internal/auth"
)

func (cfg *apiConfig) handlerLogin(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email string `json:"email"`
		Password string `json:"password"`
		ExpiresInSeconds int `json:"expires_in_seconds"`
	}

	type response struct {
		User
		Token        string `json:"token"`
		RefreshToken string `json:"refresh_token"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)

	if err != nil {
		log.Printf("Could not decode parameters: %v\n", err)
		respondWithError(w, http.StatusInternalServerError, "Internal Server Error", err)
		return
	}

	email := params.Email
	password := params.Password
	expiresInSeconds := params.ExpiresInSeconds
	expiresIn := time.Duration(1 * time.Hour)

	if !validateEmail(email) {
		respondWithError(w, http.StatusBadRequest, "Email address not formatted correctly", nil)
		return
	}

	user, err := cfg.db.GetUser(r.Context(), email)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondWithError(w, http.StatusUnauthorized, "Incorrect email or password", err)
			return
		}

		log.Printf("Could not fetch user with email: %s | %v\n", email, err)
		respondWithError(w, http.StatusInternalServerError, "Internal Server Error", err)
		return
	}

	match, err := auth.CheckPasswordHash(password, user.HashedPassword)

	if err != nil {
		log.Printf("Error checking password hash for user_id: %s | %v\n", user.ID, err)
		respondWithError(w, http.StatusInternalServerError, "Internal Server Error", err)
		return
	}

	if !match {
		respondWithError(w, http.StatusUnauthorized, "Incorrect email or password", err)
		return
	}

	if expiresInSeconds > 0 && expiresInSeconds < int((1 * time.Hour).Seconds()) {
		expiresIn = time.Duration(expiresInSeconds) * time.Second
	}

	accessToken, err := auth.MakeJWT(user.ID, cfg.jwtSecret, expiresIn)

	if err != nil {
		log.Printf("Error making JWT for user: %s | %v\n", user.ID, err)
		respondWithError(w, http.StatusInternalServerError, "Internal Server Error", err)
		return
	}

	respondWithJSON(w, http.StatusOK, response{
		User: User{
			ID: user.ID,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
			Email: user.Email,
		},
		Token: accessToken,
	})
}
