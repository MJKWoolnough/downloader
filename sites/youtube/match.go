package youtube

import "net/http"

func quickMatch(text string) bool {
	for _, r := range matches {
		if r.MatchString(text) {
			return true
		}
	}
	return false
}

func match(text string) bool {
	code := getCode(text)
	if code != "" {
		r, _ := http.Head(videoInfoURL + code)
		return r.StatusCode == http.StatusOK
	}
	return false
}
