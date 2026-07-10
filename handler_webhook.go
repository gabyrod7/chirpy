package main

import (
	"encoding/json"
	"database/sql"
	"net/http"

	"github.com/google/uuid"
	"github.com/gabyrod7/chirpy/internal/auth"
	"github.com/gabyrod7/chirpy/internal/database"
)

func (cfg *apiConfig) handlerPolkaWebhook(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Event string `json:"event"`
		Data struct {
			UserID uuid.UUID `json:"user_id"`
		} `json:"data"`
	}

	apiKey, err := auth.GetAPIKey(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't get API key", err)
		return
	}
	if apiKey != cfg.polkaKey {
		respondWithError(w, http.StatusUnauthorized, "API key from the header and environment variable do not match", err)
		return
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err = decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}
	if params.Event != "user.upgraded" {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	user, err := cfg.db.UpgradeUser(r.Context(), database.UpgradeUserParams{
		ID: params.Data.UserID,
		IsChirpyRed: sql.NullBool{Bool: true, Valid:true},
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't upgrade user", err)
		return
	}

	_, err = cfg.db.GetUserByEmail(r.Context(), user.Email)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Couldn't find user", err)
		return
	}
	
	w.WriteHeader(http.StatusNoContent)
}
