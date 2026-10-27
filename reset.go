package main

import (
	"log"
	"net/http"
)

func (cfg *apiConfig) handlerReset(w http.ResponseWriter, r *http.Request) {
	if cfg.platform != "dev" {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("403 Forbidden, only allowed in dev environment!"))
		return
	}

	err := cfg.db.DeleteAllUsers(r.Context())

	if err != nil {
		log.Printf("Could not reset database: %v\n", err)
		respondWithError(w, http.StatusInternalServerError, "Internal Server Error", err)
		return
	}

	log.Println("Users table has been reset")
	cfg.fileserverHits.Store(0)

	type returnVals struct {
		Result string `json:"result"`
	}

	respondWithJSON(w, http.StatusOK, returnVals{
		Result: "Hits reset to 0 and Database reset!",
	})
}
