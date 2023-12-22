package tests

import (
	"context"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/events/kafka"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/tests/docker"
	"reflect"
	"sync"
	"testing"
	"time"
)

func TestKafka(t *testing.T) {
	wg := &sync.WaitGroup{}
	defer wg.Wait()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	_, zkIp, err := docker.Zookeeper(ctx, wg)
	if err != nil {
		t.Fatal(err)
	}
	zookeeperUrl := zkIp + ":2181"

	kafkaUrl, err := docker.Kafka(ctx, wg, zookeeperUrl)
	if err != nil {
		t.Fatal(err)
	}

	k, err := kafka.Init(kafkaUrl, "test", true)
	if err != nil {
		t.Fatal(err)
	}

	deliveries := []string{}
	mux := sync.Mutex{}

	err = k.Consume("test", func(delivery []byte) error {
		mux.Lock()
		defer mux.Unlock()
		deliveries = append(deliveries, string(delivery))
		return nil
	})

	if err != nil {
		t.Fatal(err)
	}

	err = k.Publish("test", "a", []byte("1"))
	if err != nil {
		t.Fatal(err)
	}

	err = k.Publish("test", "b", []byte("2"))
	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(10 * time.Second)

	k.Close()

	time.Sleep(5 * time.Second)

	if !reflect.DeepEqual(deliveries, []string{"1", "2"}) {
		t.Fatal(deliveries)
	}
}
