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
	resp, err := http.Head(url)
	if err != nil {
		return nil, err
	}
	size, err := strconv.Atoi(resp.Header.Get("Content-Length"))
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	return &HTTP{
		Client:  http.DefaultClient,
		Request: req,
		Size:    int64(size),
	}, nil
}

func (h *HTTP) NewReadCloser(start, end int64) (io.ReadCloser, error) {
	h.Lock()
	defer h.Unlock()
	if start > 0 && end < h.Size {
		h.Request.Header.Add("Content-Range", strconv.Itoa(int(start))+"-"+strconv.Itoa(int(end))+"/"+strconv.Itoa(int(h.Size)))
		defer h.Request.Header.Del("Content-Range")
	}
	r, err := h.Client.Do(h.Request)
	if err != nil {
		return nil, err
	}
	if start > 0 && end < h.Size {
		if r.StatusCode != http.StatusPartialContent {
			return nil, UnexpectedStatus{r.StatusCode, http.StatusPartialContent}
		}
	} else if r.StatusCode != http.StatusOK {
		return nil, UnexpectedStatus{r.StatusCode, http.StatusOK}
	}
	return r.Body, nil
}

func (h *HTTP) Length() int64 {
	return h.Size
}

// Errors

// UnexpectedStatus is an error returned when a non-200 status is received.
type UnexpectedStatus struct {
	Got, Expected int
}

func (u UnexpectedStatus) Error() string {
	return "received status " + strconv.Itoa(u.Got) + ", expecting " + strconv.Itoa(u.Expected)
}
