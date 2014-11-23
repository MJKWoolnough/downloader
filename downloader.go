package downloader

import (
	"io"
	"time"
)

type downloader interface {
	// QuickMatch returns whether or not the downloader recognises the url,
	// unambiguosly, as one belong to it. SHOULD NOT use networking to
	// discover match.
	QuickMatch(string) bool
	// Match returns whether or not the url/string belong to this
	// downloader. This may require using the network and contacting the
	// corresponding site for confirmation.
	Match(string) bool
	// Request takes the url and returns a Request struct.
	Request(string) (*Request, error)
}

type Request struct {
	// Title is the title of the video, if it has one.
	Title string
	// Downloaders a a list of ReadClosers that all represented the requested
	// media.
	Downloaders []Media
}

type Media struct {
	// Size is the length, in bytes, of the requested media.
	Size int
	// MimeType is the mimetype of the media.
	MimeType string
	// UID is a string that uniqely identifies this request. Different
	// versions of the media (different qualities, mimetypes, etc.) should
	// have a different UID.
	UID string
	// LastModified is the last modified time of the media. Set to time.Now()
	// when undeterminable.
	LastModified time.Time
	// Sources represents a list of possible sources for this incarnation
	// of the media file
	Sources []io.ReadCloser
}

var downloaders []downloader

func Register(d downloader) {
	downloaders = append(downloaders, downloader)
}
