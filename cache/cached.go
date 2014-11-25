package cache

type CachedObject struct {
	object        *object
	filename      string
	start, length int
}
