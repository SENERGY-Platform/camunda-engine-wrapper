package kafka

import (
	"github.com/ory/dockertest"
	"reflect"
	"sync"
	"testing"
	"time"
)

func TestKafka(t *testing.T) {
	pool, err := dockertest.NewPool("")
	if err != nil {
		t.Fatal("Could not connect to docker:", err)
	}

	closeZk, _, zkIp, err := ZookeeperContainer(pool)
	defer closeZk()
	if err != nil {
		t.Fatal(err)
	}
	zookeeperUrl := zkIp + ":2181"

	//kafka
	kafkaUrl, closeKafka, err := KafkaContainer(pool, zookeeperUrl)
	defer closeKafka()
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

	time.Sleep(5 * time.Second)

	kafka.Close()

	time.Sleep(5 * time.Second)

	if !reflect.DeepEqual(deliveries, []string{"1", "2"}) {
		t.Fatal(deliveries)
	}
}
