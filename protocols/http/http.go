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
	err := h.GetLength()
	if err != nil {
		return nil, err
	}
	return h, nil
}

// GetLength sends a HEAD request in order to determine the content length
func (h *HTTP) GetLength() error {
	old := h.Request.Method
	h.Request.Method = "HEAD"
	defer func() {
		h.Request.Method = old
	}()
	resp, err := h.Client.Do(h.Request)
	if err != nil {
		return err
	}
	cl := resp.Header.Get("Content-Length")
	if len(cl) == 0 {
		return NoLength{}
	}
	size, err := strconv.Atoi(cl)
	if err != nil {
		return err
	}
	h.Size = int64(size)
	return nil
}

// NewReadCloser returns a new io.ReadCloser with the start and end bounds set
func (h *HTTP) NewReadCloser(start, end int64) (io.ReadCloser, error) {
	h.Lock()
	defer h.Unlock()
	if start > 0 {
		if end > h.Size {
			end = h.Size
		}
		h.Request.Header.Add("Range", "bytes="+strconv.Itoa(int(start))+"-"+strconv.Itoa(int(end)-1))
		defer h.Request.Header.Del("Range")
	}
	r, err := h.Client.Do(h.Request)
	if err != nil {
		return nil, err
	}
	if start > 0 {
		if r.StatusCode != http.StatusPartialContent {
			return nil, UnexpectedStatus{r.StatusCode, http.StatusPartialContent}
		}
	} else if r.StatusCode != http.StatusOK {
		return nil, UnexpectedStatus{r.StatusCode, http.StatusOK}
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
