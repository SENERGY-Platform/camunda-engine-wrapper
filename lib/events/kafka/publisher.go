package kafka

import (
	"context"
	"github.com/segmentio/kafka-go"
	"io"
	"log"
	"os"
	"runtime/debug"
	"time"
)

func (this *Kafka) Publish(topic string, key string, payload []byte) error {
	this.mux.Lock()
	defer this.mux.Unlock()
	publ, ok := this.publishers[topic]
	if !ok {
		var err error
		publ, err = NewPublisher(this.broker, topic, this.debug)
		if err != nil {
			return err
		}
		this.publishers[topic] = publ
	}
	return publ.Publish(topic, key, payload)
}

type Publisher struct {
	writer *kafka.Writer
}

func NewPublisher(broker string, topic string, debug bool) (*Publisher, error) {
	writer, err := getProducer(broker, topic, debug)
	if err != nil {
		return nil, err
	}
	return &Publisher{writer: writer}, nil
}

func (this *Publisher) Publish(topic string, key string, payload []byte) (err error) {
	err = this.writer.WriteMessages(
		context.Background(),
		kafka.Message{
			Key:   []byte(key),
			Value: payload,
			Time:  time.Now(),
		},
	)
	if err != nil {
		debug.PrintStack()
	}
	return err
}

func getProducer(broker string, topic string, debug bool) (writer *kafka.Writer, err error) {
	var logger *log.Logger
	if debug {
		logger = log.New(os.Stdout, "[KAFKA-PRODUCER] ", log.LstdFlags)
	} else {
		logger = log.New(io.Discard, "", 0)
	}
	writer = &kafka.Writer{
		Addr:        kafka.TCP(broker),
		Topic:       topic,
		MaxAttempts: 10,
		Logger:      logger,
		Async:       false,
		BatchSize:   1,
		Balancer:    &kafka.Hash{},
	}
	return writer, err
}
