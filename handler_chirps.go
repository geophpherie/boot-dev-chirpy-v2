package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/geophpherie/boot-dev-chirpy-v2/internal/auth"
	"github.com/geophpherie/boot-dev-chirpy-v2/internal/database"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerNewChirp(w http.ResponseWriter, r *http.Request) {
	requestParams := struct {
		Body string `json:"body"`
	}{}

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&requestParams); err != nil {
		log.Printf("Error decoding request body: %v", err)
		errorResponse(w, http.StatusInternalServerError, "Error decoding request body")
		return
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		errorResponse(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	userId, err := auth.ValidateJWT(token, cfg.secret)
	if err != nil {
		errorResponse(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	if len(requestParams.Body) > 140 {
		log.Printf("Chirp too long!")
		errorResponse(w, http.StatusBadRequest, "Chirp is too long")
		return
	}

	cleanChirp := badWordReplacement(requestParams.Body)

	args := database.CreateChirpParams{
		Body:   cleanChirp,
		UserID: userId,
	}

	chirp, err := cfg.dbQueries.CreateChirp(r.Context(), args)
	if err != nil {
		log.Printf("Error creating chirp! %v", err)
		errorResponse(w, http.StatusBadRequest, "Could not create chirp")
		return
	}

	response := struct {
		Id        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Body      string    `json:"body"`
		UserId    uuid.UUID `json:"user_id"`
	}{
		Id:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserId:    chirp.UserID,
	}
	jsonResponse(w, http.StatusCreated, response)

}

func (cfg *apiConfig) handlerGetAllChirps(w http.ResponseWriter, r *http.Request) {
	q_user_id := r.URL.Query().Get("author_id")
	q_sort := r.URL.Query().Get("sort")

	var chirps []database.Chirp
	var err error
	if q_user_id == "" {
		chirps, err = cfg.dbQueries.GetAllChirps(r.Context())
		if err != nil {
			log.Printf("Error getting all chirps! %v", err)
			errorResponse(w, http.StatusInternalServerError, "Could not retrieve all chirps.")
			return
		}
	} else {
		userId, err := uuid.Parse(q_user_id)
		if err != nil {
			errorResponse(w, http.StatusInternalServerError, "Unknown user")
			return

		}
		chirps, err = cfg.dbQueries.GetAllChirpsByUserId(r.Context(), userId)
		if err != nil {
			log.Printf("Error getting all chirps! %v", err)
			errorResponse(w, http.StatusInternalServerError, "Could not retrieve all chirps.")
			return
		}
	}

	type chirpResponse struct {
		Id        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Body      string    `json:"body"`
		UserId    uuid.UUID `json:"user_id"`
	}

	response := []chirpResponse{}

	// chirps come back sorted ASC, so only change to DESC if it's given
	if q_sort == "desc" {
		slices.Reverse(chirps)
	}

	for _, chirp := range chirps {
		r := chirpResponse{
			Id:        chirp.ID,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
			Body:      chirp.Body,
			UserId:    chirp.UserID,
		}
		response = append(response, r)

	}
	jsonResponse(w, http.StatusOK, response)
}

func (cfg *apiConfig) handlerGetChirp(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("chirpId"))
	if err != nil {
		log.Printf("Error getting chirp! %v", err)
		errorResponse(w, http.StatusBadRequest, "Could not use chirp id")
		return

	}
	chirp, err := cfg.dbQueries.GetChirp(r.Context(), id)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("no chirp found for id %v", id)
			errorResponse(w, http.StatusNotFound, "Chirp not found")
			return
		}
		log.Printf("Error getting chirp! %v", err)
		errorResponse(w, http.StatusInternalServerError, "Unable to retrieve chirp")
		return
	}

	type chirpResponse struct {
		Id        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Body      string    `json:"body"`
		UserId    uuid.UUID `json:"user_id"`
	}

	response := chirpResponse{
		Id:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserId:    chirp.UserID,
	}

	jsonResponse(w, http.StatusOK, response)
}

func (cfg *apiConfig) handlerDeleteChirp(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		errorResponse(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	userId, err := auth.ValidateJWT(token, cfg.secret)
	if err != nil {
		errorResponse(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	chirpId, err := uuid.Parse(r.PathValue("chirpId"))
	if err != nil {
		log.Printf("Error getting chirp! %v", err)
		errorResponse(w, http.StatusBadRequest, "Could not use chirp id")
		return

	}

	chirp, err := cfg.dbQueries.GetChirp(r.Context(), chirpId)
	if err != nil {
		log.Printf("Error getting chirp! %v", err)
		errorResponse(w, http.StatusNotFound, "Unable to retrieve chirp")
		return
	}

	if chirp.UserID != userId {
		log.Printf("Can't do that")
		errorResponse(w, http.StatusForbidden, "Unable to delete chirp")
		return

	}
	args := database.DeleteChirpParams{
		ID:     chirpId,
		UserID: userId,
	}
	err = cfg.dbQueries.DeleteChirp(r.Context(), args)
	if err != nil {
		log.Printf("Error deleting chirp! %v", err)
		errorResponse(w, http.StatusInternalServerError, "Unable to retrieve chirp")
		return
	}

	jsonResponse(w, http.StatusNoContent, struct{}{})
}

func badWordReplacement(s string) string {
	badWords := []string{"kerfuffle", "sharbert", "fornax"}
	var cleanWords []string
	for _, word := range strings.Fields(s) {
		if slices.Contains(badWords, strings.ToLower(word)) {
			log.Printf("%v is a dirty word you filthy child!", word)
			cleanWords = append(cleanWords, "****")
		} else {
			cleanWords = append(cleanWords, word)
		}
	}
	return strings.Join(cleanWords, " ")
}
