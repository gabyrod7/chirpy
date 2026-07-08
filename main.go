package main

import (
	"os"
    "fmt"
	"log"
	"database/sql"
	"net/http"
    "sync/atomic"

	"github.com/joho/godotenv"

	"github.com/gabyrod7/chirpy/internal/database"
)

import _ "github.com/lib/pq"

type apiConfig struct {
    fileserverHits atomic.Int32
	db *database.Queries
	platform string
	jwtSecret string
}

func main() {
	const filepathRoot = "."
	const port = "8080"

	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatal("DB_URL must be set")
	}

	platform := os.Getenv("PLATFORM")
	if dbURL == "" {
		log.Fatal("JWT_SECRET must be set")
	}

	jwtSecret:= os.Getenv("JWT_SECRET")
	if dbURL == "" {
		log.Fatal("JWT_SECRET must be set")
	}

	dbConn, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Error opening database: %s", err)
	}
	dbQueries := database.New(dbConn)

    apiCfg := apiConfig{
		fileserverHits: atomic.Int32{},
		db: dbQueries,
		platform: platform,
		jwtSecret: jwtSecret,
	}

	mux := http.NewServeMux()

    handler := http.StripPrefix("/app", http.FileServer(http.Dir(filepathRoot)))
    mux.Handle("/app/", apiCfg.middlewareMetricsInc(handler))
	mux.HandleFunc("GET    /api/healthz", handlerReadiness)
	mux.HandleFunc("GET    /api/chirps", apiCfg.handlerGetChirps)
	mux.HandleFunc("GET    /api/chirps/{chirpID}", apiCfg.handlerGetChirp)
	mux.HandleFunc("POST   /api/chirps", apiCfg.handlerChirps)
	mux.HandleFunc("DELETE /api/chirps/{chirpID}", apiCfg.handlerDeleteChirp)
	mux.HandleFunc("PUT  /api/users", apiCfg.handlerUsersUpdate)
	mux.HandleFunc("POST /api/users", apiCfg.handlerUsersCreate)
	mux.HandleFunc("POST /api/login", apiCfg.handlerLogin)
	mux.HandleFunc("POST /api/refresh", apiCfg.handlerRefresh)
	mux.HandleFunc("POST /api/revoke", apiCfg.handlerRevoke)
	mux.HandleFunc("GET  /admin/metrics", apiCfg.handlerCount)
	mux.HandleFunc("POST /admin/reset", apiCfg.handlerReset)

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Printf("Serving files from %s on port: %s\n", filepathRoot, port)
	log.Fatal(srv.ListenAndServe())
}

func handlerReadiness(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(http.StatusText(http.StatusOK)))
}

func (cfg *apiConfig) handlerCount(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	text := fmt.Sprintf(`<html>
  <body>
    <h1>Welcome, Chirpy Admin</h1>
    <p>Chirpy has been visited %d times!</p>
  </body>
</html>
`, cfg.fileserverHits.Load())
	w.Write([]byte(text))
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        cfg.fileserverHits.Add(1)
        next.ServeHTTP(w, r)
    })
}
