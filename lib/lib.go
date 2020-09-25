package lib

import (
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/kafka"
	"log"
)

func Wrapper() {
	cqrs, err := kafka.Init(Config.ZookeeperUrl, Config.KafkaGroup, Config.KafkaDebug)
	if err != nil {
		log.Fatal("unable to init kafka connection", err)
	}

	err = InitEventSourcing(cqrs)
	if err != nil {
		log.Fatal("unable to start eventsourcing", err)
	}

	defer CloseEventSourcing()

	InitApi()
}
