package main

import (
	"sync/atomic"

	"github.com/geophpherie/boot-dev-chirpy-v2/internal/database"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	dbQueries      database.Queries
}
