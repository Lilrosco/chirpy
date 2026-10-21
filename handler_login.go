package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/Lilrosco/chirpy/internal/auth"
)

func (cfg *apiConfig) handlerLogin(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email string `json:"email"`
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

	respondWithJSON(w, http.StatusOK, response{
		User: User{
			ID: user.ID,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
			Email: user.Email,
		},
	})
}
