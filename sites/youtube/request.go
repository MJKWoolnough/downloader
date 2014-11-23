package youtube

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"

	"github.com/MJKWoolnough/downloader"
)

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
