package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/Lilrosco/chirpy/internal/auth"
	"github.com/Lilrosco/chirpy/internal/database"
	"github.com/google/uuid"
)

var (
	accessTokenExpiresIn  = time.Hour
	refreshTokenExpiresIn = 60 * 24 * time.Hour
)

func (cfg *apiConfig) handlerLogin(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
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

	if err != nil || !match {
		respondWithError(w, http.StatusUnauthorized, "Incorrect email or password", err)
		return
	}

	accessToken, err := auth.MakeJWT(user.ID, cfg.jwtSecret, accessTokenExpiresIn)

	if err != nil {
		log.Printf("Error making JWT for user: %s | %v\n", user.ID, err)
		respondWithError(w, http.StatusInternalServerError, "Internal Server Error", err)
		return
	}

	refreshToken := auth.MakeRefreshToken()

	_, err = cfg.db.CreateRefreshToken(
		r.Context(),
		database.CreateRefreshTokenParams{
			Token:  refreshToken,
			UserID: user.ID,
			ExpiresAt: sql.NullTime{
				Time:  time.Now().Add(refreshTokenExpiresIn),
				Valid: true,
			},
			RevokedAt: sql.NullTime{},
		},
	)

	if err != nil {
		log.Printf("Could not create refresh token: %v\n", err)
		respondWithError(w, http.StatusInternalServerError, "Couldn't save refresh token", err)
		return
	}

	respondWithJSON(w, http.StatusOK, response{
		User: User{
			ID:        user.ID,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
			Email:     user.Email,
		},
		Token:        accessToken,
		RefreshToken: refreshToken,
	})
}

func (cfg *apiConfig) getUserIDFromBearerToken(w http.ResponseWriter, r *http.Request) (uuid.UUID, error) {
	token, err := auth.GetBearerToken(r.Header)

	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Could not find JWT!", err)
		return uuid.Nil, err
	}

	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)

	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized!", err)
		return uuid.Nil, err
	}

	return userID, nil
}
