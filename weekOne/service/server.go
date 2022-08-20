package service

import (
	"context"
	"net/http"
)

type ServerOption func(server *Server)

func WithBeforeCallback(rejectCallback ShutdownCallback) ServerOption {
	return func(server *Server) {
		server.beforeCallback = append(server.beforeCallback, rejectCallback)
	}
}

type Server struct {
	name           string
	serv           *http.Server
	mux            *mapHandlerRouter
	beforeCallback []ShutdownCallback
}

var _ Route = &Server{}

func NewServer(name string, addr string, opts ...ServerOption) *Server {
	router := NewMapHandlerRouter()
	s := &Server{
		name: name,
		serv: &http.Server{
			Addr:    addr,
			Handler: router,
		},
		mux:            router,
		beforeCallback: make([]ShutdownCallback, 0, 5),
	}
	for _, opt := range opts {
		opt(s)
	}

	return s
}

func (s *Server) Run() error {
	return s.serv.ListenAndServe()
}

func (s *Server) Get(path string, handler Handler) {
	s.mux.Get(path, handler)
}

func (s *Server) Put(path string, handler Handler) {
	s.mux.Put(path, handler)
}

func (s *Server) Post(path string, handler Handler) {
	s.mux.Post(path, handler)
}

func (s *Server) Delete(path string, handler Handler) {
	s.mux.Delete(path, handler)
}

func (s *Server) Use(handler Handler) {
	s.mux.Use(handler)
}

func (s *Server) Stop(ctx context.Context) error {
	return s.serv.Shutdown(ctx)
}
