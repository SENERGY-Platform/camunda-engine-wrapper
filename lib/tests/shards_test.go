package tests

import (
	"context"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/shards"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/shards/cache"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/tests/docker"
	"reflect"
	"sync"
	"testing"
)

func TestSelectShard(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	wg := sync.WaitGroup{}

	defer wg.Wait()
	defer cancel()

	pgConn, err := docker.Postgres(ctx, &wg, "test")
	if err != nil {
		t.Error(err)
		return
	}

	s, err := shards.New(pgConn, cache.None)
	if err != nil {
		t.Error(err)
		return
	}

	testSelectShard(s, t)
}

func TestSelectShardWithCache(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	wg := sync.WaitGroup{}

	defer wg.Wait()
	defer cancel()

	pgConn, err := docker.Postgres(ctx, &wg, "test")
	if err != nil {
		t.Error(err)
		return
	}

	s, err := shards.New(pgConn, cache.New(nil))
	if err != nil {
		t.Error(err)
		return
	}

	testSelectShard(s, t)

}

func testSelectShard(s *shards.Shards, t *testing.T) {
	t.Run("init shards", testInitShards(s))
	t.Run("check expected count", testCheckCount(s, map[string]int{"shard1": 0, "shard2": 1, "shard3": 2}))
	t.Run("check expected shard selection", testCheckShardSelection(s, "shard1"))
	t.Run("ensure shard for user4", testEnsureShardForUser(s, "user4", "shard1"))
	t.Run("check expected count", testCheckCount(s, map[string]int{"shard1": 1, "shard2": 1, "shard3": 2}))
	t.Run("get shard for user2", testEnsureShardForUser(s, "user2", "shard3"))
	t.Run("set shard for user5", testSetShardForUser(s, "user5", "shard2"))
	t.Run("get shard for user5", testEnsureShardForUser(s, "user5", "shard2"))
	t.Run("check expected count", testCheckCount(s, map[string]int{"shard1": 1, "shard2": 2, "shard3": 2}))
	t.Run("update shard for user2", testSetShardForUser(s, "user2", "shard2"))
	t.Run("get shard for user2 after update", testEnsureShardForUser(s, "user2", "shard2"))
	t.Run("check expected count", testCheckCount(s, map[string]int{"shard1": 1, "shard2": 3, "shard3": 1}))
}

func testSetShardForUser(s *shards.Shards, user string, shard string) func(t *testing.T) {
	return func(t *testing.T) {
		err := s.SetShardForUser(user, shard)
		if err != nil {
			t.Error(err)
			return
		}
	}
}

func testEnsureShardForUser(s *shards.Shards, user string, expectedShardUsed string) func(t *testing.T) {
	return func(t *testing.T) {
		shard, err := s.EnsureShardForUser(user)
		if err != nil {
			t.Error(err)
			return
		}
		if shard != expectedShardUsed {
			t.Error("actual:", shard, "expected:", expectedShardUsed)
			return
		}
	}
}

func testCheckCount(s *shards.Shards, expected map[string]int) func(t *testing.T) {
	return func(t *testing.T) {
		actual, err := s.GetShardUserCount()
		if err != nil {
			t.Error(err)
			return
		}
		if !reflect.DeepEqual(actual, expected) {
			t.Error("actual:", actual, "expected:", expected)
			return
		}
	}
}

func testCheckShardSelection(s *shards.Shards, expected string) func(t *testing.T) {
	return func(t *testing.T) {
		actual, err := s.SelectShard()
		if err != nil {
			t.Error(err)
			return
		}
		if actual != expected {
			t.Error("actual:", actual, "expected:", expected)
			return
		}
	}
}

func testInitShards(s *shards.Shards) func(t *testing.T) {
	return func(t *testing.T) {
		err := s.EnsureShard("shard1")
		if err != nil {
			t.Error(err)
			return
		}
		err = s.EnsureShard("shard2")
		if err != nil {
			t.Error(err)
			return
		}
		err = s.EnsureShard("shard3")
		if err != nil {
			t.Error(err)
			return
		}
		err = s.SetShardForUser("user1", "shard2")
		if err != nil {
			t.Error(err)
			return
		}

		err = s.SetShardForUser("user2", "shard3")
		if err != nil {
			t.Error(err)
			return
		}
		err = s.SetShardForUser("user3", "shard3")
		if err != nil {
			t.Error(err)
			return
		}
	}
}
