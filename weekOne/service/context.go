package service

import "net/http"

type Context struct {
	request        *http.Request
	responseWriter http.ResponseWriter
	index          int
	handle         []Handler
}

func newContext(r *http.Request, w http.ResponseWriter) *Context {
	return &Context{
		request:        r,
		responseWriter: w,
		index:          -1,
	}
}

func (c *Context) StatusCode(code int) {
	c.responseWriter.WriteHeader(code)
}

func (c *Context) Method() string {
	return c.request.Method
}
func (c *Context) Path() string {
	return c.request.URL.Path
}

func (c *Context) Writer() http.ResponseWriter {
	return c.responseWriter
}
func (c *Context) Request() *http.Request {
	return c.request
}
func (c *Context) WriteString(code int, str string) (int, error) {
	c.StatusCode(code)
	return c.responseWriter.Write([]byte(str))
}

func (c *Context) setHandler(handler ...Handler) {
	c.handle = handler
}
func (c *Context) Next() {
	c.index++
	if c.index < len(c.handle) {
		c.handle[c.index](c)
	}
}
