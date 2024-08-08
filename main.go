package main

import (
	"embed"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/shaharia-lab/headlines/headline"
)

//go:embed frontend.html
var content embed.FS

func main() {
	// Define the port flag
	port := flag.Int("port", 8080, "Port to run the server on")
	flag.Parse()

	httpClient := headline.NewCachingHTTPClient(5*time.Second, "headlines/1.0")

	sources := []headline.NewsClient{
		headline.NewProthomAloClient("https://www.prothomalo.com/", httpClient),
		headline.NewMZaminClient("https://mzamin.com/", httpClient),
		headline.NewDailyStarBanglaClient("https://bangla.thedailystar.net/", httpClient),
	}

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Compress(5))

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Content-Type"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	// Serve the index.html file for the root route
	r.Get("/", serveIndexHandler())

	r.Get("/api/headlines", headlinesHandler(sources))

	log.Printf("Starting server on :%d", *port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *port), r))
}

func serveIndexHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		indexFile, err := content.ReadFile("frontend.html")
		if err != nil {
			http.Error(w, "Could not read index file", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(indexFile)
	}
}

func headlinesHandler(sources []headline.NewsClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cachedHeadlines, isCached := headline.GetCachedHeadlines()
		if isCached {
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("X-Cache", "HIT")
			json.NewEncoder(w).Encode(cachedHeadlines)
			return
		}

		headlines := headline.GetHeadlines(sources)

		headline.CacheHeadlines(headlines)

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Cache", "MISS")
		json.NewEncoder(w).Encode(headlines)
	}
}
