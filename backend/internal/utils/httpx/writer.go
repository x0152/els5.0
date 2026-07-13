package httpx

import (
	"bufio"
	"errors"
	"net"
	"net/http"
)

type observedWriter struct {
	http.ResponseWriter
	status        int
	bytes         int64
	headerWritten bool
	hijacked      bool
}

func newObservedWriter(w http.ResponseWriter) *observedWriter {
	if ow, ok := w.(*observedWriter); ok {
		return ow
	}
	return &observedWriter{ResponseWriter: w, status: http.StatusOK}
}

func (o *observedWriter) Status() int  { return o.status }
func (o *observedWriter) Bytes() int64 { return o.bytes }
func (o *observedWriter) HeaderWritten() bool {
	return o.headerWritten
}

func (o *observedWriter) WriteHeader(status int) {
	if o.headerWritten {
		return
	}
	o.headerWritten = true
	o.status = status
	o.ResponseWriter.WriteHeader(status)
}

func (o *observedWriter) Write(p []byte) (int, error) {
	if !o.headerWritten {
		o.headerWritten = true
	}
	n, err := o.ResponseWriter.Write(p)
	o.bytes += int64(n)
	return n, err
}

func (o *observedWriter) Flush() {
	if f, ok := o.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

func (o *observedWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	h, ok := o.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, errors.New("hijack not supported")
	}
	o.headerWritten = true
	o.hijacked = true
	return h.Hijack()
}

func (o *observedWriter) Unwrap() http.ResponseWriter { return o.ResponseWriter }
