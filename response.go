package main

import (
	"encoding/json"
	"net/http"
)

func errorResponse(w http.ResponseWriter, code int, msg string) {
	error := struct {
		Error string `json:"error"`
	}{
		Error: msg,
	}

	jsonResponse(w, code, error)
}

func jsonResponse(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Add("Content-Type", "application/json")

	dat, err := json.Marshal(payload)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("unable to format response"))
	} else {
		w.WriteHeader(code)
		w.Write(dat)
	}
}
