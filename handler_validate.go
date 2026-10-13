package main

import (
	"encoding/json"
	"net/http"
)

func handlerChirpsValidation(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
        Body string `json:"body"`
    }

	type returnVals struct {
        Valid bool `json:"valid"`
    }

	decoder := json.NewDecoder(r.Body)
    params := parameters{}
    err := decoder.Decode(&params)

    if err != nil {
		respondWithError(w, http.StatusInternalServerError, "couldn't decode parameters", err)
		return
    }

	maxChirpLen := 140

	if len(params.Body) > maxChirpLen {
		respondWithError(w, http.StatusBadRequest, "Chirp too long", nil)
		return
	}

    respondWithJSON(w, http.StatusOK, returnVals{
		Valid: true,
	})
}
