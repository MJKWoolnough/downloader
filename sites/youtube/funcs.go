package youtube

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/MJKWoolnough/downloader"
	phttp "github.com/MJKWoolnough/downloader/protocols/http"
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
		regexp.MustCompile("^(?:https?://)?(?:www\\.)youtube\\.com/watch?.*v=([[[:word:]]-]{11})(?:&.*)?"),
		regexp.MustCompile("^(?:https?://)?(?:www\\.)youtube\\.com/v/([[[:word:]]-]{11})(?:\\?.*)?"),
		regexp.MustCompile("^(?:https?://)?youtu\\.be/([[[:word:]]-]{11})(\\?.*)"),
	}
	codeMatch      = regexp.MustCompile("^[[[:word:]]-]{11}$")
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

func request(text string) (*downloader.Request, error) {
	code := getCode(text)
	if code == "" {
		return nil, UnknownCode(text)
	}
	r, err := http.Get(videoInfoURL + code)
	if err != nil {
		return nil, err
	}
	if r.StatusCode != http.StatusOK {
		return nil, phttp.UnexpectedStatus(r.StatusCode)
	}
	data, err := ioutil.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		return nil, err
	}
	v, err := url.ParseQuery(string(data))
	if err != nil {
		return nil, err
	}
	for _, field := range requiredFields {
		if len(v[field]) == 0 {
			return nil, MissingField(field)
		}
	}
	streamData := strings.Split(v[fieldStreamMap][0], ",")
	streamMap := make(streams, 0, len(streamData))
StreamParseLoop:
	for _, s := range streamData {
		sm, err := url.ParseQuery(s)
		if err != nil {
			continue
		}
		for _, field := range smRequiredFields {
			if len(sm[field]) == 0 {
				continue StreamParseLoop
			}
		}
		itag, err := strconv.Atoi(sm[fieldITag][0])
		if err != nil {
			continue
		}
		q := parseQuality(sm[fieldQuality][0])
		if q == qualityUnknown {
			continue
		}
		mime := parseMime(sm[fieldMime][0])
		if mime == mimeUnknown {
			continue
		}
		u, err := url.Parse(sm[fieldURL][0])
		if err != nil {
			continue
		}
		streamMap = append(streamMap, stream{
			iTag:         itag,
			quality:      q,
			mime:         mime,
			url:          u,
			fallbackHost: sm[fieldFallbackHost][0],
		})
	}
	if len(streamMap) == 0 {
		return nil, NoStreams{}
	}
	sort.Sort(streamMap)
	media := make([]downloader.Media, 0, len(streamMap))
	for _, stream := range streamMap {
		fallback := false
		r, err := http.Head(stream.url.String())
		if err != nil {
			fallback = true
			stream.url.Host = stream.fallbackHost
			r, err = http.Head(stream.url.String())
			if err != nil {
				continue
			}
		}
		if r.StatusCode != http.StatusOK {
			continue
		}
		size, err := strconv.Atoi(r.Header.Get("Content-Length"))
		if err != nil {
			continue
		}
		lastModified, err := http.ParseTime(r.Header.Get("Last-Modified"))
		if err != nil {
			continue
		}
		sources := make([]io.ReadCloser, 0, 2)
		if !fallback {
			h, _ := phttp.NewHTTP(stream.url.String())
			sources = append(sources, h)
			stream.url.Host = stream.fallbackHost
		}
		h, _ := phttp.NewHTTP(stream.url.String())
		sources = append(sources, h)
		uid := fmt.Sprintf("youtube-%s-%d-%d-%d", code, stream.iTag, stream.quality, stream.mime)
		media = append(media, downloader.Media{
			Size:         size,
			MimeType:     stream.mime.String(),
			UID:          uid,
			LastModified: lastModified,
		})
	}
	if len(media) == 0 {
		return nil, NoStreams{}
	}
	return &downloader.Request{
		Title:       v[fieldTitle][0],
		Downloaders: media,
	}, nil
}

func getCode(text string) string {
	if codeMatch.MatchString(text) {
		return text
	} else {
		for _, r := range matches {
			s := r.FindStringSubmatch(text)
			if len(s) == 2 {
				return s[1]
			}
		}
	}
	return ""
}

// Errors

type UnknownCode string

func (u UnknownCode) Error() string {
	return "could not find youtube identifier: " + string(u)
}

type MissingField string

func (m MissingField) Error() string {
	return "could not find required field: " + string(m)
}

type NoStreams struct{}

func (NoStreams) Error() string {
	return "no valid streams found"
}
