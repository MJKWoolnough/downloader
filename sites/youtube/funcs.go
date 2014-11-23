package youtube

import (
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

const (
	fieldTitle        = "title"
	fieldStreamMap    = "url_encoded_fmt_stream_map"
	fieldITag         = "itag"
	fieldQuality      = "quality"
	fieldURL          = "url"
	fieldFallbackHost = "fallback_host"
	fieldMime         = "type"
)

var (
	matches = [...]*regexp.Regexp{
		regexp.MustCompile("^(?:https?://)?(?:www\\.)?youtube\\.com/watch?.*v=([a-zA-Z0-9_-]{11})(?:&.*)?$"),
		regexp.MustCompile("^(?:https?://)?(?:www\\.)?youtube\\.com/v/([a-zA-Z0-9_-]{11})(?:\\?.*)?$"),
		regexp.MustCompile("^(?:https?://)?youtu\\.be/([a-zA-Z0-9_-]{11})(?:\\?.*)?$"),
	}
	codeMatch      = regexp.MustCompile("^[a-zA-Z0-9_-]{11}$")
	requiredFields = [...]string{
		fieldTitle,
		fieldStreamMap,
	}
	smRequiredFields = [...]string{
		fieldITag,
		fieldQuality,
		fieldURL,
		fieldFallbackHost,
		fieldMime,
	}
)

type quality int

const (
	qualityUnknown quality = iota
	qualityHD1080
	qualitySmall
	qualityMedium
	qualityLarge
	qualityHighres
	qualityHD720
)

func parseQuality(q string) quality {
	switch q {
	case "hd1080":
		return qualityHD1080
	case "small":
		return qualitySmall
	case "medium":
		return qualityMedium
	case "large":
		return qualityMedium
	case "highres":
		return qualityHighres
	case "hd720":
		return qualityHD720
	}
	return qualityUnknown
}

type mimeType int

func (m mimeType) String() string {
	switch m {
	case mime3GPP:
		return "video/3gpp"
	case mimeFLV:
		return "video/x-flv"
	case mimeWebM:
		return "video/webm"
	case mimeMP4:
		return "video/mp4"
	}
	return "unknown mime type"
}

const (
	mimeUnknown mimeType = iota
	mime3GPP
	mimeFLV
	mimeWebM
	mimeMP4
)

func parseMime(m string) mimeType {
	if strings.HasPrefix(m, "video/3gpp") {
		return mime3GPP
	} else if strings.HasPrefix(m, "video/x-flv") {
		return mimeFLV
	} else if strings.HasPrefix(m, "video/webm") {
		return mimeWebM
	} else if strings.HasPrefix(m, "video/mp4") {
		return mimeMP4
	}
	return mimeUnknown
}

var videoInfoURL = "https://www.youtube.com/get_video_info?el=detailpage&videoid="

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

type stream struct {
	iTag int
	quality
	mime         mimeType
	url          *url.URL
	fallbackHost string
}

type streams []stream

func (s streams) Len() int {
	return len(s)
}

func (s streams) Less(i, j int) bool {
	if s[j].quality == s[i].quality {
		return s[j].mime < s[i].mime
	}
	return s[j].quality < s[i].quality
}

func (s streams) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func getCode(text string) string {
	if codeMatch.MatchString(text) {
		return text
	}
	for _, r := range matches {
		s := r.FindStringSubmatch(text)
		if len(s) == 2 {
			return s[1]
		}
	}
	return ""
}

// Errors

// UnknownCode is an error returned when no youtube identifier is found.
type UnknownCode string

func (u UnknownCode) Error() string {
	return "could not find youtube identifier: " + string(u)
}

// MissingField is an error that is returned when a required field is missing
// from the data gathered from the youtube servers.
type MissingField string

func (m MissingField) Error() string {
	return "could not find required field: " + string(m)
}

// NoStreams is an error returned when no valid streams could be found for a
// URL.
type NoStreams struct{}

func (NoStreams) Error() string {
	return "no valid streams found"
}
