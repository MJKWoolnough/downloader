package cache

import (
	"io"
	"os"
	"path"
)

type request struct {
	offset int64
	c      chan response
}

type response struct {
	size int64
	err  error
}

type object struct {
	r chan request
	q chan struct{}
}

func newObject(c *cache, key string, r io.ReadCloser) *object {
	o := &object{
		r: make(chan request),
		q: make(chan struct{}),
	}
	go o.start(c, key, rs)
	return o
}

func (o *object) Ready(size int64) (int64, error) {
	c := make(chan response)
	defer close(c)
	select {
	case <-o.q:
		return 0, ObjectRemoved{}
	case o.r <- request{size, c}:
		toRet := <-c
		return toRet.size, toRet.err
	}
}

func (o *object) start(c *cache, key string, r io.ReadCloser) {

	var total int64

	filename := path.Join(c.dir, key)

	f, err := os.Create(filename)
	if err == nil {
		requests := make([]request, 0, 8)
	CopyLoop:
		for {
			n, err = io.CopyN(f, r, 8192)
			total += n
			if err != nil {
				if err == io.EOF {
					err = nil
				} else {
					c.Remove(key)
				}
				break CopyLoop
			}
			for i := 0; i < len(requests); i++ {
				if requests[i].offset <= total {
					requests[i].c <- response{size: total}
					requests[i] = requests[len(requests)-1]
					requests = requests[:len(requests)-1]
					i--
				}
			}
		ChannelLoop1:
			for {
				select {
				case req := <-o.r:
					if req.offset <= total {
						req.c <- response{size: total}
					} else {
						requests = append(requests, req)
					}
				case <-o.q:
					break CopyLoop
				default:
					break ChannelLoop1
				}
			}
		}
		f.Close()
		r.Close()
		for _, req := range requests {
			if req.offset <= total {
				req.c <- response{size: total, err: io.EOF}
			} else {
				req.c <- response{size: total, err: err}
			}
		}
	}
ChannelLoop2:
	for {
		select {
		case req := <-o.r:
			if req.offset <= total {
				req.c <- response{size: total, err: io.EOF}
			} else {
				req.c <- response{size: total, err: err}
			}
		case <-o.q:
			os.Remove(filename)
			return
		}
	}
}

// Errors

type ObjectRemoved struct{}

func (ObjectRemoved) Error() string {
	return "object was removed from cache"
}
