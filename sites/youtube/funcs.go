package youtube

import (
	"regexp"
	"strings"
)

var (
	matches = [...]*regexp.Regexp{
		regexp.MustCompile("^(?:https?://)?(?:www\\.)?youtube\\.com/watch?.*v=([a-zA-Z0-9_-]{11})(?:&.*)?$"),
		regexp.MustCompile("^(?:https?://)?(?:www\\.)?youtube\\.com/v/([a-zA-Z0-9_-]{11})(?:\\?.*)?$"),
		regexp.MustCompile("^(?:https?://)?youtu\\.be/([a-zA-Z0-9_-]{11})(?:\\?.*)?$"),
	}
	codeMatch = regexp.MustCompile("^[a-zA-Z0-9_-]{11}$")
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
