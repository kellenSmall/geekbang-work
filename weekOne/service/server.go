package service

import (
	"context"
	"net/http"
)

type serverMux struct {
	reject bool
	*http.ServeMux
}

func (s *serverMux) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	if s.reject {
		writer.WriteHeader(http.StatusServiceUnavailable)
		_, _ = writer.Write([]byte("服务已关闭"))
		return
	}

	s.ServeMux.ServeHTTP(writer, request)
}

type Server struct {
	srv  *http.Server
	name string
	mux  *serverMux
}

func NewServer(name string, addr string) *Server {
	mux := &serverMux{ServeMux: http.NewServeMux()}
	return &Server{
		srv:  &http.Server{Addr: addr, Handler: mux},
		name: name,
		mux:  mux,
	}
}

func (s *Server) Handle(pattern string, handler http.Handler) {
	s.mux.Handle(pattern, handler)
}

func (s *Server) Start() error {
	return s.srv.ListenAndServe()
}

// RejectReq true 拒绝请求
func (s *Server) RejectReq() {
	s.mux.reject = true
}

func (s *Server) Stop(ctx context.Context) error {
	return s.srv.Shutdown(ctx)
}
