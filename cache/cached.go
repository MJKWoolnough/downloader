package cache

import (
	"io"
	"os"
)

type CachedObject struct {
	object *object
	file   *os.File
	pos    int64
}

func (c *CachedObject) Close() error {
	if c.file != nil {
		c.pos = 0
		return c.file.Close()
	}
	return nil
}

func (c *CachedObject) Read(p []byte) (int, error) {
	if c.file == nil {
		return 0, FileClosed{}
	}
	readTo := c.pos + int64(len(p))
	size, err := c.o.Ready(readTo)
	if size >= readTo {
		n, err := c.file.Read(p)
		c.pos += int64(n)
		return n, err
	}
	return 0, err
}

func (c *CachedObject) ReadAt(p []byte, off int64) (int, error) {
	if c.file == nil {
		return 0, FileClosed{}
	}
	readTo := off + int64(len(p))
	size, err := c.o.Ready(readTo)
	if size >= readTo {
		return c.file.ReadAt(p, off)
	}
	return 0, err
}

func (c *CachedObject) Seek(offset int64, whence int) (n int64, err error) {
	if c.file == nil {
		return 0, FileClosed{}
	}
	switch whence {
	case os.SEEK_SET:
		c.pos = offset
	case os.SEEK_CUR:
		c.pos += offset
	case os.SEEK_END:
		c.pos = 0
		return 0, SeekEndError{}
	default:
		return 0, UnknownWhence(whence)
	}
	if c.pos < 0 {
		c.pos = 0
		err = NegativeOffset{}
	}
	c.file.Seek(c.pos, os.SEEK_SET)
	return c.pos, err
}

func (c *CachedObject) WriterTo(w io.Writer) (int64, error) {
	if c.file == nil {
		return 0, FileClosed{}
	}
	var total int64
	for {
		size, err := c.o.Ready(c.pos + 32*1024)
		if err == io.EOF {
			n, err := io.Copy(w, c.file)
			total += n
			return total, err
		} else if size > c.pos {
			n, err := io.CopyN(w, c.file, size-c.pos)
			total += n
			c.pos += n
			if err != nil {
				return total, err
			}
		}
		if err != nil {
			return total, err
		}
	}
}

// Errors

type UnknownWhence int

func (UnknownWhence) Error() string {
	return "unknown whence"
}

type SeekEndError struct{}

func (SeekEndError) Error() string {
	return "seeking from the end is not currently supported"
}

type NegativeOffset struct{}

func (NegativeOffset) Error() string {
	return "seeked to negative offset"
}
