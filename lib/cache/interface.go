package cache

type Cache interface {
	Use(key string, getter func() (interface{}, error), result interface{}) (err error)
	Invalidate(key string) (err error)
}
