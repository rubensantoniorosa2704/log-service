package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	logCtrl "github.com/rubensantoniorosa2704/LoggingSSE/internal/infrastructure/http/controller/log"
	"github.com/rubensantoniorosa2704/LoggingSSE/internal/infrastructure/http/middleware"
	httpSwagger "github.com/swaggo/http-swagger"
)

type RouterConfig struct {
	LogController *logCtrl.LogController
	SSEServer     interface {
		Subscribe(w http.ResponseWriter, r *http.Request, applicationID string)
	}
}

func RegisterRoutes(cfg RouterConfig) http.Handler {
	r := chi.NewRouter()

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token", "X-API-Key"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	r.Use(middleware.SecurityHeaders)
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)

	r.Route("/api/v1", func(r chi.Router) {
		r.Post("/logs", cfg.LogController.CreateLogHandler)

		r.Options("/logs", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		r.Get("/events/{applicationID}", func(w http.ResponseWriter, r *http.Request) {
			cfg.SSEServer.Subscribe(w, r, chi.URLParam(r, "applicationID"))
		})
	})

	r.Handle("/docs/*", http.StripPrefix("/docs/", http.FileServer(http.Dir("docs"))))

	r.Handle("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("/docs/swagger.json"),
	))

	return r
}
