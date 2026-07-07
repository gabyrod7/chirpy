package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gabyrod7/chirpy/internal/auth"
	"github.com/gabyrod7/chirpy/internal/database"
)

type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
	Token	  string	`json"token"`
}

func (cfg *apiConfig) handlerUsersCreate(w http.ResponseWriter, r *http.Request) {
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
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}

	hashedPassword, err := auth.HashPassword(params.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't hash password", err)
		return
	}

	user, err := cfg.db.CreateUser(r.Context(), database.CreateUserParams{
		Email: params.Email,
		HashedPassword : hashedPassword,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create user", err)
		return
	}

	respondWithJSON(w, http.StatusCreated, response{
		User: User{
			ID:        user.ID,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
			Email:     user.Email,
		},
	})
}

//func (cfg *apiConfig) handlerLogin(w http.ResponseWriter, r *http.Request) {
//	type parameters struct {
//		Email string `json:"email"`
//		Password string `json:"password"`
//		ExpiresInSeconds *int `json:"expires_in_seconds"`
//	}
//	type response struct {
//		User
//	}
//
//	decoder := json.NewDecoder(r.Body)
//	params := parameters{}
//	err := decoder.Decode(&params)
//	if err != nil {
//		respondWithError(w, http.StatusBadRequest, "Couldn't decode parameters", err)
//		return
//	}
//	
//	if params.ExpiresInSeconds == nil || params.ExpiresInSeconds > 3600{
//		params.ExpiresInSeconds = 3600
//	}
//
//	user, err := cfg.db.GetUser(r.Context(), params.Email)
//	if err != nil {
//		respondWithError(w, http.StatusUnauthorized, "Incorrect email or password", err)
//		return
//	}
//
//   ok, err := auth.CheckPasswordHash(params.Password, user.HashedPassword)
//   if err != nil {
//      respondWithError(w, http.StatusInternalServerError, "Couldn't check password", err)
//      return
//   }
//
//   if !ok {
//      respondWithError(w, http.StatusUnauthorized, "Incorrect email or password", nil)
//      return
//   }
//
//   token, err := auth.GetBearerToken(r.Header)
//   if err != nil {
//      respondWithError(w, http.StatusBadRequest, "Error getting token", nil)
//      return
//   }
//
//	respondWithJSON(w, http.StatusOK, response{
//		User: User{
//			ID:        user.ID,
//			CreatedAt: user.CreatedAt,
//			UpdatedAt: user.UpdatedAt,
//			Email:     user.Email,
//			Token:	   token,
//		},
//	})
//}
