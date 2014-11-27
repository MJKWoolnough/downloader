package downloader

import (
	"io"
	"time"
)

type Site interface {
	// Match returns whether or not the url/string belong to this
	// downloader. This may require using the network and contacting the
	// corresponding site for confirmation.
	Match(string) bool
	// Request takes the url and returns a Request struct.
	Request(string) (*Request, error)
}

// Request is a value returned by a downloader and contains all of the
// necessary information to download a particular file.
type Request struct {
	// Filename is the title of the media, if it has one, and an extension.
	Filename string
	// Downloaders a a list of ReadClosers that all represented the requested
	// media.
	Downloaders []Media
}

type Downloader interface {
	NewReadCloser(int64, int64) (io.ReadCloser, error)
	Length() int64
}

// Media contains information about a particular version of a file.
type Media struct {
	// Size is the length, in bytes, of the requested media.
	Size int64
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
	Sources []Downloader
}

var sites []Site

// Register allows packages to register a Site.
func Register(s Site) {
	sites = append(sites, s)
}

func DoRequest(url string) (*Request, error) {
	for _, site := range sites {
		if site.Match(url) {
			return site.Request(url)
		}
	}
	return nil, NoRequest{}
}

// Errors

type NoRequest struct{}

func (NoRequest) Error() string {
	return "no matching request found"
}
