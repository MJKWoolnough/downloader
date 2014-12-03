package cache

import (
	"io"
	"os"
	"sync"

	"github.com/MJKWoolnough/boolmap"
	"github.com/MJKWoolnough/downloader"
	"github.com/MJKWoolnough/memio"
)

type request struct {
	startChunk, endChunk uint
	c                    chan error
}

type object struct {
	req  chan request
	quit chan struct{}
	size int64
	file *os.File
}

const chunkSize = 512 * 1024

func newObject(filename string, r downloader.Downloader) (*object, error) {
	f, err := os.Create(filename)
	if err != nil {
		return nil, err
	}
	if err = preallocate(f, r.Length()); err != nil {
		return nil, err
	}
	if err = os.Remove(filename); err != nil {
		return nil, err
	}
	o := &object{
		req:  make(chan request),
		quit: make(chan struct{}),
		size: r.Length(),
		file: f,
	}
	go o.taskMaster(r)
	return o, nil
}

func (o *object) ReadAt(b []byte, offset int64) (int, error) {
	return o.file.ReadAt(b, offset)
}

func (o *object) Size() int64 {
	return o.size
}

func (o *object) Request(start int64, length int) error {
	req := request{
		startChunk: uint(start / chunkSize),
		endChunk:   uint((start + int64(length)) / chunkSize),
		c:          make(chan error),
	}
	defer close(req.c)
	select {
	case o.req <- req:
		err := <-req.c
		return err
	case <-o.quit:
		return ObjectRemoved{}
	}
}

func (o *object) taskMaster(r downloader.Downloader) {
	size := r.Length()
	numChunks := size / chunkSize
	if size%chunkSize > 0 {
		numChunks++
	}
	ctx := &context{
		Downloader:     r,
		chunkDone:      make(chan uint),
		downloaderDone: make(chan struct{}),
		crumbslice:     boolmap.NewCrumbSliceSize(uint(numChunks)),
		numChunks:      uint(numChunks),
	}

	requests := make([]request, 0, 32)

	ctx.Set(0, 1)
	go o.download(ctx, 0)
	running := 1

downloadLoop:
	for {
		select {
		case req := <-o.req:
			for i := req.startChunk; i <= req.endChunk; i++ {
				if ctx.GetCompareSet(i, 0, 1) {
					requests = append(requests, req)
					go o.download(ctx, i)
					running++
					continue downloadLoop
				} else if ctx.Get(i) == 1 {
					requests = append(requests, req)
					continue downloadLoop
				}
			}
			req.c <- nil
		case chunk := <-ctx.chunkDone:
		checkRequestLoop:
			for i := 0; i < len(requests); i++ {
				req := requests[i]
				if req.startChunk <= chunk && req.endChunk >= chunk {
					for j := req.startChunk; j <= req.endChunk; j++ {
						if ctx.Get(j) != 2 {
							continue checkRequestLoop
						}
					}
					req.c <- nil
					requests[i] = requests[len(requests)-1]
					requests = requests[:len(requests)-1]
					i--
				}
			}
		case <-ctx.downloaderDone:
			running--
			if running == 0 {
				for i := uint(0); i <= uint(o.size/chunkSize); i++ {
					if ctx.GetCompareSet(i, 0, 1) {
						running++
						go o.download(ctx, i)
						break
					}
				}
			}
			if running == 0 {
				for _, req := range requests {
					req.c <- nil
				}
				close(ctx.downloaderDone)
				close(ctx.chunkDone)
				break downloadLoop
			}
		case <-o.quit:
			o.file.Close()
			for _, req := range requests {
				req.c <- ObjectRemoved{}
			}
			return
		}
	}
	for {
		select {
		case req := <-o.req:
			req.c <- nil
		case <-o.quit:
			o.file.Close()
			return
		}
	}
}

func (o *object) download(ctx *context, start uint) {
	defer func() {
		ctx.downloaderDone <- struct{}{}
	}()
	var end uint
	for end = start + 1; end <= ctx.numChunks; end++ {
		if ctx.Get(end) != 0 {
			break
		}
	}

	rc, err := ctx.NewReadCloser(int64(start*chunkSize), int64((start-end)*chunkSize))
	if err != nil {
		ctx.Set(start, 0)
		return
	}
	defer rc.Close()
	buf := make([]byte, chunkSize)
	w := memio.Create(&buf)
	for chunk := start; chunk < end; chunk++ {
		if !ctx.GetCompareSet(chunk, 0, 1) && chunk != start {
			return
		}
		w.Seek(0, 0)
		n, err := io.CopyN(w, rc, chunkSize)
		if err == io.EOF && chunk == ctx.numChunks-1 && n == ctx.Length()%chunkSize {
			err = nil
		} else if err != nil {
			ctx.Set(chunk, 0)
			return
		}
		_, err = o.file.WriteAt(buf[:n], int64(chunk*chunkSize))
		if err != nil {
			ctx.Set(chunk, 0)
			return
		}
		ctx.Set(chunk, 2)
		ctx.chunkDone <- chunk
	}
}

type context struct {
	downloader.Downloader
	chunkDone      chan uint
	downloaderDone chan struct{}
	crumbslice     *boolmap.CrumbSlice
	mutex          sync.RWMutex
	numChunks      uint
}

func (c *context) Get(p uint) byte {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.crumbslice.Get(p)
}

func (c *context) Set(p uint, d byte) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.crumbslice.Set(p, d)
}

func (c *context) GetCompareSet(p uint, cmp, set byte) bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	old := c.crumbslice.Get(p)
	if old == cmp {
		c.crumbslice.Set(p, set)
		return true
	}
	return false
}

// Errors

type ObjectRemoved struct{}

func (ObjectRemoved) Error() string {
	return "object was removed from cache"
}
