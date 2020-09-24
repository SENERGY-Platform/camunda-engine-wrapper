package shardmigration

import (
	"context"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/cache"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/shards"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/tests/docker"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/tests/mocks"
	"sync"
	"testing"
)

func TestMigrate(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	wg := sync.WaitGroup{}

	defer wg.Wait()
	defer cancel()

	pgConn, err := docker.Postgres(ctx, &wg)
	if err != nil {
		t.Error(err)
		return
	}

	responseSetter := &[]TenantWrapper{
		{TenantId: "t1"},
		{TenantId: "t2"},
		{TenantId: "t3"},
		{TenantId: "t1"},
	}

	camundaUrl, _ := mocks.CamundaServerWithResponse(ctx, &wg, responseSetter)

	t.Run("run migration", func(t *testing.T) {
		err = Run(camundaUrl, pgConn, 100)
		if err != nil {
			t.Error(err)
			return
		}
	})

	t.Run("check migration result", func(t *testing.T) {
		s, err := shards.New(pgConn, cache.None)
		if err != nil {
			t.Error(err)
			return
		}
		shard, err := s.GetShardForUser("t1")
		if err != nil {
			t.Error(err)
			return
		}
		if shard != camundaUrl {
			t.Error(shard, camundaUrl)
		}

		shard, err = s.GetShardForUser("t2")
		if err != nil {
			t.Error(err)
			return
		}
		if shard != camundaUrl {
			t.Error(shard, camundaUrl)
		}
		shard, err = s.EnsureShardForUser("t6")
		if err != nil {
			t.Error(err)
			return
		}
		if shard != camundaUrl {
			t.Error(shard, camundaUrl)
		}

	})
}
