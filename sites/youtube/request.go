package youtube

import (
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"

	"github.com/MJKWoolnough/downloader"
	phttp "github.com/MJKWoolnough/downloader/protocols/http"
)

const (
	fieldTitle        = "title"
	fieldStreamMap    = "url_encoded_fmt_stream_map"
	fieldQuality      = "quality"
	fieldURL          = "url"
	fieldFallbackHost = "fallback_host"
	fieldMime         = "type"
)

var (
	videoInfoURL   = "https://www.youtube.com/get_video_info?el=detailpage&video_id="
	requiredFields = [...]string{
		fieldTitle,
		fieldStreamMap,
	}
	smRequiredFields = [...]string{
		fieldQuality,
		fieldURL,
		fieldFallbackHost,
		fieldMime,
	}
)

func getYoutubeData(u string) (url.Values, error) {
	r, err := http.Get(u)
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
	return v, nil
}

type stream struct {
	quality
	mime         mimeType
	url          *url.URL
	fallbackHost string
}

type streams []*stream

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

func validateStreamData(s string) *stream {
	sm, err := url.ParseQuery(s)
	if err != nil {
		return nil
	}
	for _, field := range smRequiredFields {
		if len(sm[field]) == 0 {
			return nil
		}
	}
	q := parseQuality(sm[fieldQuality][0])
	if q == qualityUnknown {
		return nil
	}
	mime := parseMime(sm[fieldMime][0])
	if mime == mimeUnknown {
		return nil
	}
	u, err := url.Parse(sm[fieldURL][0])
	if err != nil {
		return nil
	}
	return &stream{
		quality:      q,
		mime:         mime,
		url:          u,
		fallbackHost: sm[fieldFallbackHost][0],
	}
}

func streamParser(s *stream, code string) *downloader.Media {
	fallback := false
	r, err := http.Head(s.url.String())
	if err != nil {
		fallback = true
		s.url.Host = s.fallbackHost
		r, err = http.Head(s.url.String())
		if err != nil {
			return nil
		}
	}
	if r.StatusCode != http.StatusOK {
		return nil
	}
	size, err := strconv.Atoi(r.Header.Get("Content-Length"))
	if err != nil {
		return nil
	}
	lastModified, err := http.ParseTime(r.Header.Get("Last-Modified"))
	if err != nil {
		return nil
	}
	sources := make([]io.ReadCloser, 0, 2)
	if !fallback {
		h, _ := phttp.NewHTTP(s.url.String())
		sources = append(sources, h)
		s.url.Host = s.fallbackHost
	}
	h, _ := phttp.NewHTTP(s.url.String())
	sources = append(sources, h)
	uid := "youtube-" + code + "-" + strconv.Itoa(int(s.quality)) + "-" + strconv.Itoa(int(s.mime))
	return &downloader.Media{
		Size:         size,
		MimeType:     s.mime.String(),
		UID:          uid,
		LastModified: lastModified,
		Sources:      sources,
	}
}

func request(text string) (*downloader.Request, error) {
	code := getCode(text)
	if code == "" {
		return nil, UnknownCode(text)
	}
	v, err := getYoutubeData(videoInfoURL + code)
	if err != nil {
		return nil, err
	}
	streamData := strings.Split(v[fieldStreamMap][0], ",")
	streamMap := make(streams, 0, len(streamData))
	for _, s := range streamData {
		streamInfo := validateStreamData(s)
		if streamInfo == nil {
			continue
		}
		streamMap = append(streamMap, streamInfo)
	}
	if len(streamMap) == 0 {
		return nil, NoStreams{}
	}
	sort.Sort(streamMap)
	media := make([]downloader.Media, 0, len(streamMap))
	for _, stream := range streamMap {
		m := streamParser(stream, code)
		if m == nil {
			continue
		}
		media = append(media, *m)
	}
	if len(media) == 0 {
		return nil, NoStreams{}
	}
	return &downloader.Request{
		Filename:    v[fieldTitle][0] + ".mp4",
		Downloaders: media,
	}, nil
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
