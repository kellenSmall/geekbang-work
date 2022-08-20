package service

import (
	"net/http"
)

type Route interface {
	Get(path string, handler Handler)
	Put(path string, handler Handler)
	Post(path string, handler Handler)
	Delete(path string, handler Handler)
	Use(handler Handler)
}

var _ Route = &mapHandlerRouter{}

type Handler func(ctx *Context)

type mapHandlerRouter struct {
	routers     map[string]map[string]Handler
	middlewares []Handler
}

func NewMapHandlerRouter() *mapHandlerRouter {
	routers := make(map[string]map[string]Handler, 4)
	routers[http.MethodGet] = make(map[string]Handler, 5)
	routers[http.MethodPost] = make(map[string]Handler, 5)
	routers[http.MethodPut] = make(map[string]Handler, 5)
	routers[http.MethodDelete] = make(map[string]Handler, 5)
	return &mapHandlerRouter{
		routers:     routers,
		middlewares: make([]Handler, 0, 5),
	}
}

func (m *mapHandlerRouter) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	ctx := newContext(request, writer)
	h, ok := m.routers[ctx.Method()][ctx.Path()]
	if !ok {
		_, _ = ctx.WriteString(http.StatusNotFound, "404 Not Found")
		return
	}
	handlers := append(m.middlewares, h)
	ctx.setHandler(handlers...)
	ctx.Next()
}

func (m *mapHandlerRouter) addRouter(method string, path string, handle Handler) {
	m.routers[method][path] = handle
}

func (m *mapHandlerRouter) Get(path string, handler Handler) {
	m.addRouter(http.MethodGet, path, handler)
}

func (m *mapHandlerRouter) Put(path string, handler Handler) {
	m.addRouter(http.MethodPut, path, handler)

}

func (m *mapHandlerRouter) Post(path string, handler Handler) {
	m.addRouter(http.MethodPost, path, handler)

}

func (m *mapHandlerRouter) Delete(path string, handler Handler) {
	m.addRouter(http.MethodDelete, path, handler)

}

func (m *mapHandlerRouter) Use(handler Handler) {
	m.middlewares = append(m.middlewares, handler)
}
