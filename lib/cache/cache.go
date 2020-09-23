package cache

import (
	"encoding/json"
	"errors"
	"github.com/bradfitz/gomemcache/memcache"
	"github.com/coocood/freecache"
	"log"
	"sync"
)

var DefaultL1Expiration = 10          //10sec
var DefaultL1Size = 100 * 1024 * 1024 //100MB
var DefaultL2Expiration int32 = 300   //60sec

type LayeredCache struct {
	l1     *freecache.Cache
	l2     *memcache.Client
	mux    sync.Mutex
	Debug  bool
	config *CacheConfig
}

type Item struct {
	Key   string
	Value []byte
}

type CacheConfig struct {
	L1Expiration   int
	L1Size         int
	L2Expiration   int32
	L2MemcacheUrls []string
}

var ErrNotFound = errors.New("key not found in cache")

func New(config *CacheConfig) (result *LayeredCache) {
	if config == nil {
		config = &CacheConfig{}
	}
	if config.L1Expiration == 0 {
		config.L1Expiration = DefaultL1Expiration
	}
	if config.L1Size == 0 {
		config.L1Size = DefaultL1Size
	}
	if config.L2Expiration == 0 {
		config.L2Expiration = DefaultL2Expiration
	}
	result = &LayeredCache{config: config, l1: freecache.NewCache(config.L1Size)}
	if len(config.L2MemcacheUrls) > 0 {
		result.l2 = memcache.New(config.L2MemcacheUrls...)
	}
	return
}

func (this *LayeredCache) Use(key string, getter func() (interface{}, error), result interface{}) (err error) {
	item, err := this.Get(key)
	if err == nil {
		err = json.Unmarshal(item.Value, result)
		return
	}
	temp, err := getter()
	if err != nil {
		return err
	}
	value, err := json.Marshal(temp)
	if err != nil {
		return err
	}
	this.Set(key, value)
	return json.Unmarshal(value, &result)
}

func (this *LayeredCache) Invalidate(key string) (err error) {
	this.l1.Del([]byte(key))
	if this.l2 != nil {
		err = this.l2.Delete(key)
	}
	return
}

func (this *LayeredCache) Get(key string) (item Item, err error) {
	this.mux.Lock()
	defer this.mux.Unlock()
	item.Value, err = this.l1.Get([]byte(key))
	if err != nil && err != freecache.ErrNotFound {
		log.Println("ERROR: in LayeredCache::l1.Get()", err)
	}
	if err != nil && this.l2 != nil {
		if this.Debug {
			log.Println("DEBUG: use l2 cache", key, err)
		}
		var temp *memcache.Item
		temp, err = this.l2.Get(key)
		if err == memcache.ErrCacheMiss {
			err = ErrNotFound
			return
		}
		if err != nil {
			return
		}
		err := this.l1.Set([]byte(key), temp.Value, this.config.L1Expiration)
		if err != nil {
			log.Println("ERROR: in LayeredCache::l1.Set()", err)
		}
		item.Value = temp.Value
	}
	return
}

func (this *LayeredCache) Set(key string, value []byte) {
	this.mux.Lock()
	defer this.mux.Unlock()
	err := this.l1.Set([]byte(key), value, this.config.L1Expiration)
	if err != nil {
		log.Println("ERROR: in LayeredCache::l1.Set()", err)
	}
	if this.l2 != nil {
		err = this.l2.Set(&memcache.Item{Value: value, Expiration: this.config.L2Expiration, Key: key})
		if err != nil {
			log.Println("ERROR: in LayeredCache::l2.Set()", err)
		}
	}
	return
}
