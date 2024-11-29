package main

import (
	"net/http"
)

func main() {
	mux := http.NewServeMux()

	config := apiConfig{}

	fsHandler := config.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir("."))))
	mux.Handle("/app/", fsHandler)

	mux.HandleFunc("GET /admin/metrics", config.handlerMetrics)
	mux.HandleFunc("POST /admin/reset", config.handlerReset)

	mux.HandleFunc("GET /api/healthz", handlerReadiness)
	mux.HandleFunc("POST /api/validate_chirp", handlerValidate)

	server := http.Server{
		Handler: mux,
		Addr:    ":8080"}

	err := server.ListenAndServe()
	if err != nil {
		panic(err)
	}
}
