package http

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	logCtrl "github.com/rubensantoniorosa2704/LoggingSSE/internal/infrastructure/http/controller/log"
	httpSwagger "github.com/swaggo/http-swagger"
)

const (
	APIVersion     = "/api/v1"
	LogsEndpoint   = "/logs"
	EventsEndpoint = "/events/{applicationID}"
	DocsPath       = "/docs/*"
	SwaggerPath    = "/swagger/*"
)

type RouterConfig struct {
	LogController *logCtrl.LogController
	SSEServer     interface {
		HTTPHandler(http.ResponseWriter, *http.Request)
	}
}

func RegisterRoutes(cfg RouterConfig) http.Handler {
	r := newBaseRouter()
	r.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	r.Route(APIVersion, func(r chi.Router) {
		registerLogRoutes(r, cfg)
	})

	registerDocsRoutes(r)
	return r
}

func newBaseRouter() chi.Router {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(corsHandler())
	return r
}

func corsHandler() func(http.Handler) http.Handler {
	return cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token", "X-API-Key"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	})
}

func registerLogRoutes(r chi.Router, cfg RouterConfig) {
	r.Post(LogsEndpoint, cfg.LogController.CreateLogHandler)
	r.Options(LogsEndpoint, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	r.Get(EventsEndpoint, sseHandler(http.HandlerFunc(cfg.SSEServer.HTTPHandler)))
}

func sseHandler(sse http.Handler) http.HandlerFunc {
	type contextKey string
	const applicationIDKey contextKey = "applicationID"

	return func(w http.ResponseWriter, req *http.Request) {
		applicationID := chi.URLParam(req, "applicationID")
		ctx := context.WithValue(req.Context(), applicationIDKey, applicationID)
		sse.ServeHTTP(w, req.WithContext(ctx))
	}
}

func registerDocsRoutes(r chi.Router) {
	r.Handle(DocsPath, http.StripPrefix("/docs/", http.FileServer(http.Dir("docs"))))
	r.Handle(SwaggerPath, httpSwagger.Handler(
		httpSwagger.URL("/docs/swagger.json"),
	))
}
