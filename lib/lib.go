package lib

import (
	"context"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/api"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/camunda"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/configuration"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/events"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/events/kafka"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/shards"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/shards/cache"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/vid"
)

func Wrapper(parentCtx context.Context, config configuration.Config) (err error) {
	ctx, cancel := context.WithCancel(parentCtx)
	defer func() {
		if err != nil {
			cancel()
		}
	}()

	v, err := vid.New(config.WrapperDb)
	if err != nil {
		return err
	}

	s, err := shards.New(config.ShardingDb, cache.New(&cache.CacheConfig{L1Expiration: 60}))
	if err != nil {
		return err
	}

	c := camunda.New(config, v, s)

	cqrs, err := kafka.Init(config.KafkaUrl, config.KafkaGroup, config.Debug)
	if err != nil {
		return err
	}

	go func() {
		<-ctx.Done()
		cqrs.Close()
	}()

	e, err := events.New(config, cqrs, v, c)
	if err != nil {
		return err
	}

	err = api.Start(ctx, config, c, e)
	if err != nil {
		return
	}

	return nil
}
