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
	return &Server{server: r3sse.New()}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	streamName := r.URL.Query().Get("stream")
	if streamName == "" {
		http.Error(w, "missing 'stream' query parameter", http.StatusBadRequest)
		return
	}

	// Only create stream if it doesn't exist to avoid redundant operations
	if !s.server.StreamExists(streamName) {
		s.server.CreateStream(streamName)
		log.Printf("[SSE] Channel=%s event=created", streamName)
	}

	log.Printf("[SSE] Channel=%s event=client_connected", streamName)

	go func() {
		<-ctx.Done()
		log.Printf("[SSE] Channel=%s event=client_disconnected", streamName)
	}()

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
