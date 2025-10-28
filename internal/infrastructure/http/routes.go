package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	logCtrl "github.com/rubensantoniorosa2704/LoggingSSE/internal/infrastructure/http/controller/log"
	httpSwagger "github.com/swaggo/http-swagger"
)

type RouterConfig struct {
	LogController *logCtrl.LogController
	SSEServer     interface {
		HTTPHandler(http.ResponseWriter, *http.Request)
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

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Route("/api/v1", func(r chi.Router) {
		// Log routes
		r.Post("/logs", cfg.LogController.CreateLogHandler)

		// OPTIONS for CORS preflight
		r.Options("/logs", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		// SSE route for log events by applicationID
		r.Get("/events/{applicationID}", func(w http.ResponseWriter, req *http.Request) {
			applicationID := chi.URLParam(req, "applicationID")
			// Pass as ?stream=applicationID for the SSE lib
			q := req.URL.Query()
			q.Set("stream", applicationID)
			req.URL.RawQuery = q.Encode()
			cfg.SSEServer.HTTPHandler(w, req)
		})
	})

	r.Handle("/docs/*", http.StripPrefix("/docs/", http.FileServer(http.Dir("docs"))))

	r.Handle("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("/docs/swagger.json"),
	))

	return r
}
