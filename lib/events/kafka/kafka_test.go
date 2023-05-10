package kafka

import (
	"context"
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

	_, zkIp, err := ZookeeperContainer(ctx, wg)
	if err != nil {
		t.Fatal(err)
	}
	zookeeperUrl := zkIp + ":2181"

	//kafka
	kafkaUrl, err := KafkaContainer(ctx, wg, zookeeperUrl)
	if err != nil {
		t.Fatal(err)
	}

	kafka, err := Init(kafkaUrl, "test", true)
	if err != nil {
		t.Fatal(err)
	}

	deliveries := []string{}
	mux := sync.Mutex{}

	err = kafka.Consume("test", func(delivery []byte) error {
		mux.Lock()
		defer mux.Unlock()
		deliveries = append(deliveries, string(delivery))
		return nil
	})

	if err != nil {
		t.Fatal(err)
	}

	err = kafka.Publish("test", "a", []byte("1"))
	if err != nil {
		t.Fatal(err)
	}

	err = kafka.Publish("test", "b", []byte("2"))
	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(10 * time.Second)

	kafka.Close()

	time.Sleep(5 * time.Second)

	if !reflect.DeepEqual(deliveries, []string{"1", "2"}) {
		t.Fatal(deliveries)
	}
}
