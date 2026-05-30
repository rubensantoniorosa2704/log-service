package sse

import (
	"log"
	"net/http"

	r3sse "github.com/r3labs/sse/v2"
)

type Server struct {
	server *r3sse.Server
}

func NewServer() *Server {
	s := r3sse.New()
	return &Server{server: s}
}

func (s *Server) HTTPHandler(w http.ResponseWriter, r *http.Request) {
	streamName := r.URL.Query().Get("stream")
	if streamName == "" {
		http.Error(w, "Application ID (stream) query parameter is required.", http.StatusBadRequest)
		return
	}

	s.server.CreateStream(streamName)
	log.Printf("New client connected to SSE channel (ApplicationID): %s", streamName)
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
