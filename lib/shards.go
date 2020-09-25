package lib

import (
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/cache"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/shards"
	"sync"
)

var shardsInstance *shards.Shards
var shardsOnce sync.Once

func GetShards() (instance *shards.Shards, err error) {
	shardsOnce.Do(func() {
		shardsInstance, err = shards.New(Config.PgConn, cache.New(&cache.CacheConfig{
			L1Expiration: 600, //10 min
		}))
		if err != nil {
			return
		}
	})
	return shardsInstance, err
}

func GetUserShard(userId string) (shardUrl string, err error) {
	s, err := GetShards()
	if err != nil {
		return shardUrl, err
	}
	return s.EnsureShardForUser(userId)
}
