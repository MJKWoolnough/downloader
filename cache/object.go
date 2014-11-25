package cache

import (
	"io"
	"os"
	"path"
	"sync"
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
	sync.RWMutex
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

func (o *object) Ready(size int64) error {
	o.RLock()
	select {
	case <-o.q:
		o.RUnlock()
		return ObjectRemoved{}
	default:
	}
	c := make(chan response)
	defer close(c)
	o.r <- request{size, c}
	o.RUnlock()
	return <-c
}

func (o *object) start(c *cache, key string, r io.ReadCloser) {
	buf := make([]byte, 8192)
	var total int64

	f, err := os.Create(path.Join(c.dir, filename))
	if err != nil {
		o.Lock()
		defer o.Unlock()
		o.err = err
		o.r = nil
		return
	}

	requests := make([]request, 0, 8)

	for {
		n, err = io.CopyN(f, r, 8192)
		total += n
		if err != nil {
			if err == io.EOF {
				err = nil
			} else {
				c.Remove(key)
			}
			break
		}
		for i := 0; i < len(requests); i++ {
			if requests[i].offset <= total {
				requests[i].c <- response{size: total}
				requests[i] = requests[len(requests)-1]
				requests = requests[:len(requests)-1]
				i--
			}
		}
		for {
			select {
			case req := <-o.r:
				if req.offset <= total {
					req.c <- response{size: total}
				} else {
					requests = append(requests, req)
				}
			default:
			}
		}
		select {
		case <-o.q:
			break
		default:
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
	requests = nil
	for {
		select {
		case req := <-o.r:
			if req.offset <= total {
				req.c <- response{size: total, err: io.EOF}
			} else {
				req.c <- response{size: total, err: err}
			}
		case <-o.q:
			o.Lock()
			o.r = nil
			o.Unlock()
			break
		}
	}
}

// Errors

type ObjectRemoved struct{}

func (ObjectRemoved) Error() string {
	return "object was removed from cache"
}
