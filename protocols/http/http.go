package http

import (
	"io"
	"net/http"
	"strconv"
	"sync"
)

// HTTP turns an http request into a io.ReadCloser.
type HTTP struct {
	Client  *http.Client
	Request *http.Request
	Size    int64
	sync.Mutex
}

// NewHTTP is a constructor for a simple http GET request. For a more complex
// construction, use the struct directly.
func NewHTTP(url string) (*HTTP, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	h := &HTTP{
		Client:  http.DefaultClient,
		Request: req,
	}
	err = h.GetLength()
	if err != nil {
		return nil, err
	}
	return h, nil
}

// GetLength sends a HEAD request in order to determine the content length
func (h *HTTP) GetLength() error {
	old := h.Request.Method
	h.Request.Method = "HEAD"
	resp, err := h.Client.Do(h.Request)
	h.Request.Method = old
	if err != nil {
		return err
	}
	if len(resp.Header.Get("Content-Length")) == 0 {
		return NoLength{}
	}
	h.Size = resp.ContentLength
	return nil
}

// NewReadCloser returns a new io.ReadCloser with the start and end bounds set
func (h *HTTP) NewReadCloser(start, length int64) (io.ReadCloser, error) {
	h.Lock()
	defer h.Unlock()
	if length < 0 {
		length = h.Size
	}
	if start+length > h.Size {
		length = h.Size - start
	}
	expecting := http.StatusOK
	if start > 0 || length != h.Size {
		h.Request.Header.Add("Range", "bytes="+strconv.Itoa(int(start))+"-"+strconv.Itoa(int(start+length-1)))
		defer h.Request.Header.Del("Range")
		expecting = http.StatusPartialContent
	}
	r, err := h.Client.Do(h.Request)
	if err != nil {
		return nil, err
	}
	if r.StatusCode != expecting {
		return nil, UnexpectedStatus{r.StatusCode, expecting}
	}
	return r.Body, nil
}

// Length returns the total length of the request
func (h *HTTP) Length() int64 {
	return h.Size
}

// Errors

// NoLength is an error returned when the length of the resource could not be
// automatically determined
type NoLength struct{}

func (NoLength) Error() string {
	return "could not determine length"
}

// UnexpectedStatus is an error returned when a non-200 status is received.
type UnexpectedStatus struct {
	Got, Expected int
}

func (u UnexpectedStatus) Error() string {
	return "received status " + strconv.Itoa(u.Got) + ", expecting " + strconv.Itoa(u.Expected)
}
