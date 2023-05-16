package middles

import (
	"bytes"
	"io"
	"net/http"
)

type ResponseWithRecorder struct {
	http.ResponseWriter
	statusCode int
	body       bytes.Buffer
}

type Middleware struct {
	beforeFn   func(w *ResponseWithRecorder)
	afterFn    func(w *ResponseWithRecorder, r *http.Request)
	Param      []byte
	StatusCode int
	Body       []byte
}

func NewMiddleware() *Middleware {
	return &Middleware{}
}

func (m *Middleware) Before(f func(w *ResponseWithRecorder)) {
	m.beforeFn = f
}

func (m *Middleware) After(f func(w *ResponseWithRecorder, r *http.Request)) {
	m.afterFn = f
}

func (m *Middleware) Serve(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rwr := &ResponseWithRecorder{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
			body:           bytes.Buffer{},
		}

		//after
		defer func() {
			if m.afterFn != nil {
				m.StatusCode = rwr.statusCode
				m.Body = rwr.body.Bytes()
				m.afterFn(rwr, r)
			}
		}()

		//before
		if m.beforeFn != nil {
			m.beforeFn(rwr)
		}

		//body
		m.Param, _ = io.ReadAll(r.Body)
		r.Body = io.NopCloser(bytes.NewBuffer(m.Param))

		// Call the next handler, which can be another middleware in the chain, or the final handler.
		next.ServeHTTP(rwr, r)
	})
}

func (rwr *ResponseWithRecorder) Header() http.Header {
	return rwr.ResponseWriter.Header()
}

func (rwr *ResponseWithRecorder) Write(d []byte) (int, error) {
	n, err := rwr.ResponseWriter.Write(d)
	if err != nil {
		return n, err
	}
	rwr.body.Write(d)
	return n, err
}

func (rwr *ResponseWithRecorder) WriteHeader(statusCode int) {
	rwr.ResponseWriter.WriteHeader(statusCode)
	rwr.statusCode = statusCode
}
