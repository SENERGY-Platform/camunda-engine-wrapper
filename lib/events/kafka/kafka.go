package kafka

import (
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/events/kafka/topicconfig"
	"github.com/wvanbergen/kazoo-go"
	"io/ioutil"
	"log"
	"sync"
)

type Kafka struct {
	mux        sync.Mutex
	zk         string
	group      string
	broker     []string
	consumers  []*Consumer
	publishers map[string]*Publisher
	debug      bool
}

func Init(zookeeperUrl string, group string, debug bool) (Interface, error) {
	k := Kafka{zk: zookeeperUrl, group: group, debug: debug, publishers: map[string]*Publisher{}}
	var err error
	k.broker, err = GetBroker(zookeeperUrl)
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

func GetBroker(zk string) (brokers []string, err error) {
	return getBroker(zk)
}

func getBroker(zkUrl string) (brokers []string, err error) {
	zookeeper := kazoo.NewConfig()
	zookeeper.Logger = log.New(ioutil.Discard, "", 0)
	zk, chroot := kazoo.ParseConnectionString(zkUrl)
	zookeeper.Chroot = chroot
	if kz, err := kazoo.NewKazoo(zk, zookeeper); err != nil {
		return brokers, err
	} else {
		return kz.BrokerList()
	}
}

func InitTopic(zkUrl string, topics ...string) (err error) {
	for _, topic := range topics {
		err = topicconfig.EnsureWithZk(zkUrl, topic, map[string]string{
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
