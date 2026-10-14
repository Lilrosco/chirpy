package main

import (
	"encoding/json"
	"net/http"
	"strings"
)

func handlerChirpsValidation(w http.ResponseWriter, r *http.Request) {
	badWords := map[string]string {
		"kerfuffle" : "kerfuffle",
		"sharbert" : "sharbert",
		"fornax" : "fornax",
	}

	type parameters struct {
        Body string `json:"body"`
    }

	type returnVals struct {
        CleanedBody string `json:"cleaned_body"`
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

	words := strings.Split(params.Body, " ")

	for index, word := range words {
		_, exists := badWords[strings.ToLower(word)]

		if exists {
			words[index] = "****"
		}
	}

    respondWithJSON(w, http.StatusOK, returnVals{
		CleanedBody: strings.Join(words, " "),
	})
}
