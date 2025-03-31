package server

import (
	"context"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/api"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/camunda"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/configuration"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/controller"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/metrics"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/shards"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/shards/cache"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/tests/docker"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/vid"
	"net/http"
	"net/http/httptest"
	"sync"
	"time"
)

func CreateTestEnv(ctx context.Context, wg *sync.WaitGroup, initConf configuration.Config) (config configuration.Config, wrapperUrl string, shard string, err error) {
	config = initConf
	incidentApiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	go func() {
		<-ctx.Done()
		incidentApiServer.Close()
	}()
	config.IncidentApiUrl = incidentApiServer.URL

	pgStr, err := docker.Postgres(ctx, wg, "vid_relations")
	if err != nil {
		return config, wrapperUrl, shard, err
	}

	_, camundaPgIp, _, err := docker.PostgresWithNetwork(ctx, wg, "camunda")
	if err != nil {
		return config, wrapperUrl, shard, err
	}

	camundaUrl, err := docker.Camunda(ctx, wg, camundaPgIp, "5432")
	if err != nil {
		return config, wrapperUrl, shard, err
	}

	config.WrapperDb = pgStr
	config.ShardingDb = pgStr

	s, err := shards.New(config.ShardingDb, cache.None)
	if err != nil {
		return config, wrapperUrl, shard, err
	}
	err = s.EnsureShard(camundaUrl)
	if err != nil {
		return config, wrapperUrl, shard, err
	}
	shard = camundaUrl

	v, err := vid.New(config.WrapperDb)
	if err != nil {
		return config, wrapperUrl, shard, err
	}

	c := camunda.New(config, v, s, nil)

	ctrl := controller.New(config, c, v, nil)

	httpServer := httptest.NewServer(api.GetRouter(config, c, ctrl, metrics.New()))
	wg.Add(1)
	go func() {
		<-ctx.Done()
		httpServer.Close()
		wg.Done()
	}()
	wrapperUrl = httpServer.URL
	time.Sleep(time.Second)
	return config, wrapperUrl, shard, nil
}
