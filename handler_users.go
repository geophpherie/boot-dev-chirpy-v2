package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/geophpherie/boot-dev-chirpy-v2/internal/auth"
	"github.com/geophpherie/boot-dev-chirpy-v2/internal/database"
	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID `json:"id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Email        string    `json:"email"`
	Token        string    `json:"token,omitempty"`
	RefreshToken string    `json:"refresh_token,omitempty"`
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

	expiresIn := time.Duration(60) * time.Second

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

	token, err := auth.MakeJWT(user.ID, cfg.secret, expiresIn)
	if err != nil {
		errorResponse(w, http.StatusUnauthorized, "Incorrect email or password")
		return
	}

	refreshToken, err := auth.MakeRefreshToken()
	if err != nil {
		errorResponse(w, http.StatusUnauthorized, "Incorrect email or password")
		return
	}

	// need to store token
	expiration := time.Now().Add(time.Duration(60) * time.Duration(24) * time.Hour)
	args := database.CreateRefreshTokenParams{
		Token:     refreshToken,
		CreatedAt: sql.NullTime{Time: time.Now(), Valid: true},
		UpdatedAt: sql.NullTime{Time: time.Now(), Valid: true},
		UserID:    user.ID,
		ExpiresAt: sql.NullTime{Time: expiration, Valid: true},
		RevokedAt: sql.NullTime{Valid: false},
	}
	cfg.dbQueries.CreateRefreshToken(r.Context(), args)

	responseUser := User{
		ID:           user.ID,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
		Email:        user.Email,
		Token:        token,
		RefreshToken: refreshToken,
	}

	jsonResponse(w, http.StatusOK, responseUser)
}

func (cfg *apiConfig) handlerRefresh(w http.ResponseWriter, r *http.Request) {
	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		errorResponse(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	userId, err := cfg.dbQueries.GetUserFromRefreshToken(r.Context(), refreshToken)
	if err != nil {
		errorResponse(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	expiresIn := time.Duration(60) * time.Second
	newToken, err := auth.MakeJWT(userId, cfg.secret, expiresIn)
	if err != nil {
		errorResponse(w, http.StatusUnauthorized, "Incorrect email or password")
		return
	}

	jsonResponse(w, http.StatusOK, struct {
		Token string `json:"token"`
	}{Token: newToken})

}

func (cfg *apiConfig) handlerRevoke(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		// Handle error reading the body
		fmt.Println("a")
		http.Error(w, "Request body is empty", http.StatusBadRequest)
		return
	}

	if len(body) != 0 {
		// Body is empty
		errorResponse(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		fmt.Println("a")
		errorResponse(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	err = cfg.dbQueries.RevokeRefreshToken(r.Context(), refreshToken)
	if err != nil {
		errorResponse(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	jsonResponse(w, http.StatusNoContent, struct{}{})

}
