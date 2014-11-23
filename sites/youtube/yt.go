// +build !testing

// Package youtube implements a Youtube downloader
package youtube

import "github.com/MJKWoolnough/downloader"

func init() {
	downloader.Register(new(youtube))
}

type youtube struct{}

func (youtube) QuickMatch(text string) bool {
	return quickMatch(text)
}

func (youtube) Match(text string) bool {
	return match(text)
}

func (youtube) Request(text string) (*downloader.Request, error) {
	return request(text)
}
