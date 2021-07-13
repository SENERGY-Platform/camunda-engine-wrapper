package shardmigration

import (
	"context"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/shards"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/shards/cache"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/tests/docker"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/tests/mocks"
	"reflect"
	"sync"
	"testing"
)

func TestMigrate(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	wg := sync.WaitGroup{}

	defer wg.Wait()
	defer cancel()

	pgConn, err := docker.Postgres(ctx, &wg, "test")
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

	camundaUrl2, _ := mocks.CamundaServerWithResponse(ctx, &wg, responseSetter)

	t.Run("empty remove", func(t *testing.T) {
		err = Remove(camundaUrl, pgConn)
		if err != nil {
			t.Error(err)
			return
		}
	})

	t.Run("run add", func(t *testing.T) {
		err = Add(camundaUrl, pgConn, 100)
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

		shards, err := s.GetShards()
		if err != nil {
			t.Error(err)
			return
		}
		if !reflect.DeepEqual(shards, []string{camundaUrl}) {
			t.Error(shards)
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

	t.Run("run remove", func(t *testing.T) {
		err = Remove(camundaUrl, pgConn)
		if err != nil {
			t.Error(err)
			return
		}
	})

	t.Run("check remove result", func(t *testing.T) {
		s, err := shards.New(pgConn, cache.None)
		if err != nil {
			t.Error(err)
			return
		}
		shards, err := s.GetShards()
		if err != nil {
			t.Error(err)
			return
		}

		if len(shards) != 0 {
			t.Error(shards)
			return
		}
	})

	t.Run("run add 2", func(t *testing.T) {
		err = Add(camundaUrl2, pgConn, 100)
		if err != nil {
			t.Error(err)
			return
		}
	})

	t.Run("check migration 2 result", func(t *testing.T) {
		s, err := shards.New(pgConn, cache.None)
		if err != nil {
			t.Error(err)
			return
		}

		shards, err := s.GetShards()
		if err != nil {
			t.Error(err)
			return
		}
		if !reflect.DeepEqual(shards, []string{camundaUrl2}) {
			t.Error(shards)
			return
		}

		shard, err := s.GetShardForUser("t1")
		if err != nil {
			t.Error(err)
			return
		}
		if shard != camundaUrl2 {
			t.Error(shard, camundaUrl2)
		}

		shard, err = s.GetShardForUser("t2")
		if err != nil {
			t.Error(err)
			return
		}
		if shard != camundaUrl2 {
			t.Error(shard, camundaUrl2)
		}
		shard, err = s.EnsureShardForUser("t6")
		if err != nil {
			t.Error(err)
			return
		}
		if shard != camundaUrl2 {
			t.Error(shard, camundaUrl2)
		}
	})
}
