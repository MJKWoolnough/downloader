package http

import (
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestGetLength(t *testing.T) {
	var lengthStr string
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if lengthStr != "" {
			w.Header().Add("Content-Length", lengthStr)
		}
	}))

	s.Config.ErrorLog = log.New(ioutil.Discard, "", 0)
	tests := []struct {
		lengthStr string
		length    int64
		err       string
	}{
		{"5", 5, ""},
		{"25", 25, ""},
		{"a", 0, "Head " + s.URL + ": bad Content-Length \"a\""},
		{"-1", 0, "Head " + s.URL + ": bad Content-Length \"-1\""},
	}

	for n, test := range tests {
		lengthStr = test.lengthStr
		h, err := NewHTTP(s.URL)
		errStr := ""
		if err != nil {
			errStr = err.Error()
		}
		if errStr != test.err {
			t.Errorf("test %d: expecting error %s, got %s", n+1, test.err, errStr)
		} else if test.length > 0 {
			if h == nil {
				t.Errorf("test %d: error - received nil HTTP", n+1)
			} else if h.Size != test.length {
				t.Errorf("test %d: expecting size %d, got %d", n+1, test.length, h.Size)
			}
		} else if h != nil {
			t.Errorf("test %d: error - expecting nil HTTP", n+1)
		}
	}
}

func TestNewReadCloser(t *testing.T) {
	data := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	dataReader := strings.NewReader(data)

	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeContent(w, r, "data.txt", time.Now(), dataReader)
	}))

	h, err := NewHTTP(s.URL)
	if err != nil {
		t.Errorf("unexpected error: %s", err)
		return
	}

	tests := []struct {
		start, length int64
	}{
		{0, 10},
		{1, 10},
		{2, 20},
		{20, 40},
		{20, 10},
	}

	for n, test := range tests {
		r, err := h.NewReadCloser(test.start, test.length)
		if err != nil {
			t.Errorf("test %d: unexpected error - %s", n+1, err)
		} else if d, err := ioutil.ReadAll(r); err != nil {
			t.Errorf("test %d: unexpected error - %s", n+1, err)
		} else if string(d) != data[test.start:test.start+test.length] {
			t.Errorf("test %d: expecting %s, got %s", n+1, data[test.start:test.start+test.length], d)
		}
		if r != nil {
			r.Close()
		}
	}
}
