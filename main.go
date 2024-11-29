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

	config := apiConfig{dbQueries: *dbQueries}

	mux := http.NewServeMux()

	fsHandler := config.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir("."))))
	mux.Handle("/app/", fsHandler)

	mux.HandleFunc("GET /admin/metrics", config.handlerMetrics)
	mux.HandleFunc("POST /admin/reset", config.handlerReset)

	mux.HandleFunc("GET /api/healthz", handlerReadiness)
	mux.HandleFunc("POST /api/validate_chirp", handlerValidate)

	server := http.Server{
		Handler: mux,
		Addr:    ":8080"}

	err = server.ListenAndServe()
	if err != nil {
		panic(err)
	}
}
