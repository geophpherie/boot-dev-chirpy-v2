package main

import (
	"database/sql"
	"net/http"
	"os"

	"github.com/geophpherie/boot-dev-chirpy-v2/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	godotenv.Load()

	dbUrl := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbUrl)
	if err != nil {
		panic(err)
	}

	dbQueries := database.New(db)

	config := apiConfig{
		dbQueries: *dbQueries,
		secret:    os.Getenv("SECRET"),
	}

	mux := http.NewServeMux()

	fsHandler := config.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir("."))))
	mux.Handle("/app/", fsHandler)

	mux.HandleFunc("GET /admin/metrics", config.handlerMetrics)
	mux.HandleFunc("POST /admin/reset", config.handlerReset)

	mux.HandleFunc("GET /api/healthz", handlerReadiness)

	mux.HandleFunc("POST /api/chirps", config.handlerNewChirp)
	mux.HandleFunc("GET /api/chirps", config.handlerGetAllChirps)
	mux.HandleFunc("GET /api/chirps/{chirpId}", config.handlerGetChirp)
	mux.HandleFunc("DELETE /api/chirps/{chirpId}", config.handlerDeleteChirp)

	mux.HandleFunc("POST /api/users", config.handlerNewUser)
	mux.HandleFunc("PUT /api/users", config.handlerUpdateUser)
	mux.HandleFunc("POST /api/login", config.handlerLogin)
	mux.HandleFunc("POST /api/refresh", config.handlerRefresh)
	mux.HandleFunc("POST /api/revoke", config.handlerRevoke)

	server := http.Server{
		Handler: mux,
		Addr:    ":8080"}

	err = server.ListenAndServe()
	if err != nil {
		panic(err)
	}
}
