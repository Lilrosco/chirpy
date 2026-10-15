package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/mail"
	"time"

	"github.com/lib/pq"
	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}

func validateEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}

func (cfg *apiConfig) handlerCreateUser(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
        Email string `json:"email"`
    }

	type response struct {
		User
	}

	decoder := json.NewDecoder(r.Body)
    params := parameters{}
    err := decoder.Decode(&params)

    if err != nil {
		respondWithError(w, http.StatusInternalServerError, "couldn't decode parameters", err)
		return
    }

	email := params.Email

	if !validateEmail(email) {
		respondWithError(w, http.StatusBadRequest, "email address not formatted correctly", nil)
		return
	}

	user, err := cfg.db.CreateUser(r.Context(), email)

	if pqErr, ok := errors.AsType[*pq.Error](err); ok {
		if pqErr.Code == errUniqueViolation {
			respondWithError(w, http.StatusBadRequest, fmt.Sprintf("User with email: %s already exists\n", email), nil)
			return
		} else if err != nil {
			respondWithError(w, http.StatusInternalServerError, "could not create user", err)
			return
		}
	}

	respondWithJSON(w, http.StatusCreated, response{
		User: User{
			ID: user.ID,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
			Email: user.Email,
		},
	})
}
