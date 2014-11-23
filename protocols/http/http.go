package http

import (
	"io"
	"net/http"
	"strconv"
)

type HTTP struct {
	Client  *http.Client
	Request *http.Request
	data    io.ReadCloser
}

// NewHTTP is a constructor for a simple http GET request. For a more complex
// construction, use the struct directly.
func NewHTTP(url string) (*HTTP, error) {
	r, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	return &HTTP{
		Client:  http.DefaultClient,
		Request: r,
	}, nil
}

func (h *HTTP) Read(d []byte) (int, error) {
	if h.data == nil {
		r, err := h.Client.Do(h.Request)
		if err != nil {
			return 0, err
		}
		if r.StatusCode != http.StatusOK {
			return 0, UnexpectedStatus(r.StatusCode)
		}
		h.data = r.Body
	}
	return h.data.Read(d)
}

func (h *HTTP) Close() error {
	if h.data == nil {
		return nil
	}
	err := h.data.Close()
	h.data = nil
	return err
}

// Errors

type UnexpectedStatus int

func (u UnexpectedStatus) Error() string {
	return "received non-200 status: " + strconv.Itoa(int(u))
}
