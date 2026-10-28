package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/mail"
	"time"

	"github.com/Lilrosco/chirpy/internal/auth"
	"github.com/Lilrosco/chirpy/internal/database"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

type User struct {
	ID          uuid.UUID `json:"id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Email       string    `json:"email"`
	IsChirpyRed bool      `json:"is_chirpy_red"`
}

func validateEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}

func (cfg *apiConfig) handlerCreateUser(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	type response struct {
		User
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

	hashed_password, err := auth.HashPassword(password)

	if err != nil {
		log.Printf("Could not hash password for email: %s | %v\n", email, err)
		respondWithError(w, http.StatusInternalServerError, "Internal Server Error", err)
		return
	}

	user, err := cfg.db.CreateUser(
		r.Context(),
		database.CreateUserParams{
			Email:          email,
			HashedPassword: hashed_password,
		},
	)

	if pqErr, ok := errors.AsType[*pq.Error](err); ok {
		if pqErr.Code == errUniqueViolation {
			respondWithError(w, http.StatusBadRequest, fmt.Sprintf("User with email: %s already exists\n", email), nil)
			return
		} else if err != nil {
			log.Printf("Could not create user: %v\n", err)
			respondWithError(w, http.StatusInternalServerError, "Internal Server Error", err)
			return
		}
	}

	respondWithJSON(w, http.StatusCreated, response{
		User: User{
			ID:          user.ID,
			CreatedAt:   user.CreatedAt,
			UpdatedAt:   user.UpdatedAt,
			Email:       user.Email,
			IsChirpyRed: user.IsChirpyRed,
		},
	})
}

func (cfg *apiConfig) handlerUpdateUser(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	type response struct {
		User
	}

	token, err := auth.GetBearerToken(r.Header)

	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Missing authorization token!", err)
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

	email := params.Email
	password := params.Password

	if !validateEmail(email) {
		respondWithError(w, http.StatusBadRequest, "Email address not formatted correctly", nil)
		return
	}

	hashed_password, err := auth.HashPassword(password)

	if err != nil {
		log.Printf("Could not hash password for email: %s | %v\n", email, err)
		respondWithError(w, http.StatusInternalServerError, "Internal Server Error", err)
		return
	}

	user, err := cfg.db.UpdateUser(
		r.Context(),
		database.UpdateUserParams{
			Email:          email,
			HashedPassword: hashed_password,
			ID:             userID,
		},
	)

	if err != nil {
		log.Printf("Could not update user: %v\n", err)
		respondWithError(w, http.StatusInternalServerError, "Internal Server Error", err)
		return
	}

	respondWithJSON(w, http.StatusOK, response{
		User: User{
			ID:          user.ID,
			CreatedAt:   user.CreatedAt,
			UpdatedAt:   user.UpdatedAt,
			Email:       user.Email,
			IsChirpyRed: user.IsChirpyRed,
		},
	})
}
