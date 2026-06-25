package main

import (
    "fmt"
	"log"
	"encoding/json"
	"net/http"
    "sync/atomic"
)

type apiConfig struct {
    fileserverHits atomic.Int32
}

func main() {
	const filepathRoot = "."
	const port = "8080"

	mux := http.NewServeMux()

    apiCfg := apiConfig{fileserverHits: atomic.Int32{},}
    handler := http.StripPrefix("/app", http.FileServer(http.Dir(filepathRoot)))
    mux.Handle("/app/", apiCfg.middlewareMetricsInc(handler))
	mux.HandleFunc("GET /api/healthz", handlerReadiness)
	mux.HandleFunc("POST /api/validate_chirp", handlerValidChirp)
	mux.HandleFunc("GET /admin/metrics", apiCfg.handlerCount)
	mux.HandleFunc("POST /admin/reset", apiCfg.handlerCountReset)

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
    //w.Write([]byte(fmt.Sprintf("Hits: %d", cfg.fileserverHits.Load())))
	text := fmt.Sprintf(`<html>
  <body>
    <h1>Welcome, Chirpy Admin</h1>
    <p>Chirpy has been visited %d times!</p>
  </body>
</html>
`, cfg.fileserverHits.Load())
	w.Write([]byte(text))
}

func (cfg *apiConfig) handlerCountReset(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
    cfg.fileserverHits.Store(0)
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        cfg.fileserverHits.Add(1)
        next.ServeHTTP(w, r)
    })
}

func handlerValidChirp(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body string `json:body"`
	}
	type errorResponse struct {
		Error string `json:"error"`
	}
	type validResponse struct {
		Valid bool `json:"valid"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		log.Printf("Error decoding parameters: %s", err)
		//w.WriteHeader(http.StatusBadRequest)
		return
	}

	if len(params.Body) > 140 {
		resp := errorResponse{
			Error: "Chirp is too long",
		}
		dat, err := json.Marshal(resp)
		if err != nil {
			w.WriteHeader(500)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write(dat)
		return
	}

	resp := validResponse{
		Valid: true,
	}
	dat, err := json.Marshal(resp)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(dat)
}
