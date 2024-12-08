package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/geophpherie/boot-dev-chirpy-v2/internal/auth"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerWebhooks(w http.ResponseWriter, r *http.Request) {
	apiKey, err := auth.GetAPIKey(r.Header)
	if err != nil {
		errorResponse(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	if apiKey != os.Getenv("POLKA_KEY") {
		errorResponse(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	requestParams := struct {
		Event string `json:"event"`
		Data  struct {
			UserId string `json:"user_id"`
		} `json:"data"`
	}{}

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&requestParams); err != nil {
		log.Printf("Error decoding request body: %v", err)
		errorResponse(w, http.StatusInternalServerError, "Error decoding request body")
		return
	}

	if requestParams.Event != "user.upgraded" {
		jsonResponse(w, http.StatusNoContent, struct{}{})
		return
	}

	userId, err := uuid.Parse(requestParams.Data.UserId)
	if err != nil {
		errorResponse(w, http.StatusInternalServerError, "Error decoding user Id")
		return
	}

	err = cfg.dbQueries.UpgradeUser(r.Context(), userId)
	if err != nil {
		errorResponse(w, http.StatusNotFound, "")
		return
	}

	jsonResponse(w, http.StatusNoContent, struct{}{})
}
