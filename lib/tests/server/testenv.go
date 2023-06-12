package server

import (
	"context"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/api"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/camunda"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/configuration"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/events"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/metrics"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/shards"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/shards/cache"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/tests/docker"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/tests/mocks"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/vid"
	"net/http/httptest"
	"sync"
)

func CreateTestEnv(ctx context.Context, wg *sync.WaitGroup, initConf configuration.Config) (config configuration.Config, wrapperUrl string, shard string, e *events.Events, err error) {
	config = initConf
	pgStr, err := docker.Postgres(ctx, wg, "vid_relations")
	if err != nil {
		return config, wrapperUrl, shard, e, err
	}

	_, camundaPgIp, _, err := docker.PostgresWithNetwork(ctx, wg, "camunda")
	if err != nil {
		return config, wrapperUrl, shard, e, err
	}

	camundaUrl, err := docker.Camunda(ctx, wg, camundaPgIp, "5432")
	if err != nil {
		return config, wrapperUrl, shard, e, err
	}

	config.WrapperDb = pgStr
	config.ShardingDb = pgStr

	s, err := shards.New(config.ShardingDb, cache.None)
	if err != nil {
		return config, wrapperUrl, shard, e, err
	}
	err = s.EnsureShard(camundaUrl)
	if err != nil {
		return config, wrapperUrl, shard, e, err
	}
	shard = camundaUrl

	v, err := vid.New(config.WrapperDb)
	if err != nil {
		return config, wrapperUrl, shard, e, err
	}

	c := camunda.New(config, v, s, nil)

	e, err = events.New(config, mocks.Kafka(), v, c, nil)
	if err != nil {
		return config, wrapperUrl, shard, e, err
	}

	httpServer := httptest.NewServer(api.GetRouter(config, c, e, metrics.New()))
	wg.Add(1)
	go func() {
		<-ctx.Done()
		httpServer.Close()
		wg.Done()
	}()
	wrapperUrl = httpServer.URL
	return config, wrapperUrl, shard, e, nil
}
