package sse

import (
	"fmt"
	"log"
	"net/http"
	"sync"
)

type subscriber chan []byte

type Server struct {
	mu          sync.RWMutex
	subscribers map[string][]subscriber
}

func NewServer() *Server {
	return &Server{subscribers: make(map[string][]subscriber)}
}

// Subscribe handles an SSE connection for the given applicationID.
func (s *Server) Subscribe(w http.ResponseWriter, r *http.Request, applicationID string) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	ch := make(subscriber, 16)
	s.add(applicationID, ch)
	defer s.remove(applicationID, ch)

	log.Printf("SSE client connected: %s", applicationID)

	for {
		select {
		case <-r.Context().Done():
			log.Printf("SSE client disconnected: %s", applicationID)
			return
		case data, ok := <-ch:
			if !ok {
				return
			}
			fmt.Fprintf(w, "data: %s\n\n", data)
			flusher.Flush()
		}
	}
}

// Publish sends data to all subscribers of the given channel.
func (s *Server) Publish(channel string, data []byte) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, ch := range s.subscribers[channel] {
		select {
		case ch <- data:
		default:
		}
	}
}

// StreamExists reports whether there are active subscribers for the channel.
func (s *Server) StreamExists(channel string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.subscribers[channel]) > 0
}

func (s *Server) add(channel string, ch subscriber) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.subscribers[channel] = append(s.subscribers[channel], ch)
}

func (s *Server) remove(channel string, ch subscriber) {
	s.mu.Lock()
	defer s.mu.Unlock()
	subs := s.subscribers[channel]
	for i, sub := range subs {
		if sub == ch {
			s.subscribers[channel] = append(subs[:i], subs[i+1:]...)
			break
		}
	}
	if len(s.subscribers[channel]) == 0 {
		delete(s.subscribers, channel)
	}
	close(ch)
}
