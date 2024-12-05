package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/geophpherie/boot-dev-chirpy-v2/internal/auth"
	"github.com/geophpherie/boot-dev-chirpy-v2/internal/database"
	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}

func (cfg *apiConfig) handlerNewUser(w http.ResponseWriter, r *http.Request) {
	requestParams := struct {
		Password string `json:"password"`
		Email    string `json:"email"`
	}{}

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&requestParams); err != nil {
		log.Printf("Error decoding request body: %v", err)
		errorResponse(w, http.StatusBadRequest, "email is not usable")
		return
	}

	hashedPassword, err := auth.HashPassword(requestParams.Password)
	if err != nil {
		errorResponse(w, http.StatusInternalServerError, "Cannot handle that password")
		return
	}

	params := database.CreateUserParams{
		Email:          requestParams.Email,
		HashedPassword: sql.NullString{String: hashedPassword, Valid: true},
	}

	user, err := cfg.dbQueries.CreateUser(r.Context(), params)
	if err != nil {
		log.Printf("User creation failed :: %v", err)
		errorResponse(w, http.StatusInternalServerError, "Cannot create user")
		return
	}

	responseUser := User{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
	}

	jsonResponse(w, http.StatusCreated, responseUser)
}

func (cfg *apiConfig) handlerLogin(w http.ResponseWriter, r *http.Request) {
	requestParams := struct {
		Password string `json:"password"`
		Email    string `json:"email"`
	}{}

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&requestParams); err != nil {
		log.Printf("Error decoding request body: %v", err)
		errorResponse(w, http.StatusBadRequest, "email is not usable")
		return
	}

	user, err := cfg.dbQueries.GetUserByEmail(r.Context(), requestParams.Email)
	if err != nil {
		errorResponse(w, http.StatusUnauthorized, "Incorrect email or password")
		return
	}

	if !user.HashedPassword.Valid {
		errorResponse(w, http.StatusUnauthorized, "Incorrect email or password")
		return
	}

	err = auth.CheckPasswordHash(requestParams.Password, user.HashedPassword.String)
	if err != nil {
		errorResponse(w, http.StatusUnauthorized, "Incorrect email or password")
		return
	}

	responseUser := User{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
	}

	jsonResponse(w, http.StatusOK, responseUser)
}
