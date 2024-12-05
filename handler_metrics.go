package main

import (
	"fmt"
	"net/http"
	"os"
)

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) handlerMetrics(w http.ResponseWriter, _ *http.Request) {
	w.Header().Add("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	html := fmt.Sprintf(`
		<html>
			<body>
				<h1>Welcome, Chirpy Admin</h1>
				<p>Chirpy has been visited %d times!</p>
			</body>

		</html>`,
		cfg.fileserverHits.Load(),
	)
	w.Write([]byte(html))
}

func (cfg *apiConfig) handlerReset(w http.ResponseWriter, r *http.Request) {
	if os.Getenv("PLATFORM") != "dev" {
		jsonResponse(w, http.StatusForbidden, "")
		return
	}
	cfg.fileserverHits.Store(0)

	err := cfg.dbQueries.DeleteAllUsers(r.Context())
	if err != nil {
		errorResponse(w, http.StatusInternalServerError, "Could not delete all users!")
		return
	}

	w.Header().Add("Content-Type", "text/plain: charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("Hits: %v", cfg.fileserverHits.Load())))
}
