package youtube

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"testing"
)

func urlValues(data ...string) url.Values {
	v := make(url.Values)
	for _, datum := range data {
		d := strings.SplitN(datum, "=", 2)
		if len(d) != 2 {
			continue
		}
		v.Set(d[0], d[1])
	}
	return v
}

type stringHTTP string

func (s stringHTTP) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(s))
}

func TestGetYoutubeData(t *testing.T) {
	tests := []struct {
		data string
		url.Values
		err error
	}{
		{"", nil, MissingField(fieldTitle)},
		{urlValues(fieldTitle + "=Hello World").Encode(), nil, MissingField(fieldStreamMap)},
		{
			urlValues(fieldTitle+"=Hello World", fieldStreamMap+"=Test1").Encode(),
			urlValues(fieldTitle+"=Hello World", fieldStreamMap+"=Test1"),
			nil,
		},
	}

	var s stringHTTP

	srv := httptest.NewServer(&s)

	for n, test := range tests {
		s = stringHTTP(test.data)
		v, err := getYoutubeData(srv.URL)
		if !reflect.DeepEqual(err, test.err) {
			t.Errorf("test %d: expecting error %s, got %s", n+1, test.err, err)
		} else if !reflect.DeepEqual(v, test.Values) {
			t.Errorf("test %d: values not identical", n+1)
		}
	}
}

func TestValidataStreamData(t *testing.T) {
	tests := []struct {
		data    string
		success bool
	}{
		{"", false},
		{urlValues(fieldQuality + "=highres").Encode(), false},
		{urlValues(fieldQuality+"=highres", fieldMime+"=video/webm").Encode(), false},
		{urlValues(fieldQuality+"=highres", fieldMime+"=video/webm", fieldURL+"=http://www.youtube.com/").Encode(), false},
		{urlValues(fieldQuality+"=highres", fieldMime+"=video/webm", fieldURL+"=http://www.youtube.com/", fieldFallbackHost+"=google.com").Encode(), true},
		{urlValues(fieldQuality+"=unknown", fieldMime+"=video/webm", fieldURL+"=http://www.youtube.com/", fieldFallbackHost+"=google.com").Encode(), false},
		{urlValues(fieldQuality+"=highres", fieldMime+"=video/unknown", fieldURL+"=http://www.youtube.com/", fieldFallbackHost+"=google.com").Encode(), false},
	}

	for n, test := range tests {
		if (validateStreamData(test.data) == nil) == test.success {
			t.Errorf("test %d: expecting sucess %v, got %v", n+1, test.success, !test.success)
		}
	}
}
