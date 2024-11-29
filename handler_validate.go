package main

import (
	"encoding/json"
	"log"
	"net/http"
	"slices"
	"strings"
)

func handlerValidate(w http.ResponseWriter, r *http.Request) {
	requestParams := struct {
		Body string `json:"body"`
	}{}

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&requestParams); err != nil {
		log.Printf("Error decoding request body: %v", err)
		errorResponse(w, http.StatusInternalServerError, "Error decoding request body")
		return
	}

	if len(requestParams.Body) > 140 {
		log.Printf("Chirp too long!")
		errorResponse(w, http.StatusBadRequest, "Chirp is too long")
		return
	}

	cleanChirp := badWordReplacement(requestParams.Body)

	response := struct {
		CleanedBody string `json:"cleaned_body"`
	}{
		CleanedBody: cleanChirp,
	}
	jsonResponse(w, http.StatusOK, response)
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
