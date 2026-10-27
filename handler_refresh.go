package main

import (
	"net/http"
	"time"

	"github.com/Lilrosco/chirpy/internal/auth"
)

func (cfg *apiConfig) handlerRefreshToken(w http.ResponseWriter, r *http.Request) {
	type response struct {
		Token string `json:"token"`
	}

	token, err := auth.GetBearerToken(r.Header)

	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Missing authorization token!", err)
		return
	}

	refreshToken, err := cfg.db.GetRefreshTokenWithUser(r.Context(), token)

	if err != nil || refreshToken.RevokedAt.Valid || refreshToken.ExpiresAt.Time.Before(time.Now()) {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized!", err)
		return
	}

	accessToken, err := auth.MakeJWT(refreshToken.UserID, cfg.jwtSecret, accessTokenExpiresIn)

	respondWithJSON(w, http.StatusOK, response{
		Token: accessToken,
	})
}

func (cfg *apiConfig) handlerRevokeToken(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)

	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Could not find token", err)
		return
	}

	err = cfg.db.RevokeRefreshToken(r.Context(), token)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not revoke session", err)
		return
	}

	respondWithJSON(w, http.StatusNoContent, nil)
}
