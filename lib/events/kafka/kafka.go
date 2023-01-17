package kafka

import (
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/events/kafka/topicconfig"
	"github.com/segmentio/kafka-go"
	"log"
	"net"
	"strconv"
	"sync"
)

type Kafka struct {
	mux               sync.Mutex
	kafkaBootstrapUrl string
	group             string
	broker            []string
	consumers         []*Consumer
	publishers        map[string]*Publisher
	debug             bool
}

func Init(kafkaBootstrapUrl string, group string, debug bool) (Interface, error) {
	k := Kafka{kafkaBootstrapUrl: kafkaBootstrapUrl, group: group, debug: debug, publishers: map[string]*Publisher{}}
	var err error
	k.broker, err = GetBroker(kafkaBootstrapUrl)
	return &k, err
}

func (this *Kafka) Close() {
	this.mux.Lock()
	defer this.mux.Unlock()
	for _, c := range this.consumers {
		c.Stop()
	}
	for _, c := range this.publishers {
		err := c.writer.Close()
		if err != nil {
			log.Println(err)
		}
	}
}

func (this *Kafka) EnsureTopic(bootstrapUrl string, topic string, config map[string]string) (err error) {
	return topicconfig.Ensure(bootstrapUrl, topic, config)
}

func GetBroker(bootstrapUrl string) (brokers []string, err error) {
	return getBroker(bootstrapUrl)
}

func getBroker(bootstrapUrl string) (result []string, err error) {
	conn, err := kafka.Dial("tcp", bootstrapUrl)
	if err != nil {
		return result, err
	}
	defer conn.Close()
	brokers, err := conn.Brokers()
	if err != nil {
		return result, err
	}
	for _, broker := range brokers {
		result = append(result, net.JoinHostPort(broker.Host, strconv.Itoa(broker.Port)))
	}
	return result, nil
}

func InitTopic(bootstrapUrl string, topics ...string) (err error) {
	for _, topic := range topics {
		err = topicconfig.Ensure(bootstrapUrl, topic, map[string]string{
			"retention.ms":              "-1",
			"retention.bytes":           "-1",
			"cleanup.policy":            "compact",
			"delete.retention.ms":       "86400000",
			"segment.ms":                "604800000",
			"min.cleanable.dirty.ratio": "0.1",
		})
		if err != nil {
			return err
		}
	}
	return nil
}
