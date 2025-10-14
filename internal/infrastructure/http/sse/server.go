package sse

import (
	"log"
	"net/http"
	"strings"

	r3sse "github.com/r3labs/sse/v2"
)

type Server struct {
	server *r3sse.Server
}

func NewServer() *Server {
	s := r3sse.New()
	// Server configs stay here
	// Example: s.Headers = map[string]string{"Access-Control-Allow-Origin": "*"}
	return &Server{server: s}
}

func (s *Server) HTTPHandler(w http.ResponseWriter, r *http.Request) {
	streamName := strings.TrimPrefix(r.URL.Path, "/events/")
	if streamName == "" {
		http.Error(w, "Channel name (e.g., /events/app-id) is required.", http.StatusBadRequest)
		return
	}

	q := r.URL.Query()
	q.Set("stream", streamName)
	r.URL.RawQuery = q.Encode()

	log.Printf("New client connected to SSE channel: %s", streamName)
	s.server.ServeHTTP(w, r)
}

func (s *Server) Publish(channel string, data []byte) {
	s.server.Publish(channel, &r3sse.Event{
		Data: data,
	})
}

func (s *Server) StreamExists(channel string) bool {
	return s.server.StreamExists(channel)
}

func (s *Server) Close() {
	s.server.Close()
}
