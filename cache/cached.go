package cache

import "io"

type CachedObject struct {
	o   *object
	pos int64
}

func (c *CachedObject) Read(p []byte) (int, error) {
	n, err := c.ReadAt(p, c.pos)
	c.pos += int64(n)
	return n, err
}

func (c *CachedObject) ReadAt(p []byte, off int64) (int, error) {
	if err := c.o.Request(off, len(p)); err != nil {
		return 0, err
	}
	return c.o.file.ReadAt(p, off)
}

func (c *CachedObject) Seek(offset int64, whence int) (int64, error) {
	switch whence {
	case 0:
		c.pos = offset
	case 1:
		c.pos += offset
	case 2:
		c.pos = c.o.size + offset
	default:
		c.pos = 0
		return 0, UnknownWhence(whence)
	}
	if c.pos < 0 {
		c.pos = 0
		return 0, NegativeOffset{}
	}
	return c.pos, nil
}

func (c *CachedObject) WriterTo(w io.Writer) (int64, error) {
	var (
		read int64
		err  error
		n    int
	)
	buf := make([]byte, 32*1024)
	for c.pos < c.o.size {
		err = c.o.Request(c.pos, len(buf))
		if err != nil {
			break
		}
		n, err = c.o.file.Read(buf)
		c.pos += int64(n)
		_, e := w.Write(buf[:n])
		if err != nil {
			break
		} else if e != nil {
			err = e
			break
		}
	}
	return read, err
}

// Errors

type UnknownWhence int

func (UnknownWhence) Error() string {
	return "unknown whence"
}

type NegativeOffset struct{}

func (NegativeOffset) Error() string {
	return "can't seek to negative offset"
}
