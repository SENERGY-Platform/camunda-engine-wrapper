/*
 * Copyright 2019 InfAI (CC SES)
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package kafka

import (
	"context"
	"github.com/segmentio/kafka-go"
	"io"
	"io/ioutil"
	"log"
	"runtime/debug"
	"sync"
	"time"
)

func (this *Kafka) Consume(topic string, listener func(delivery []byte) error) (err error) {
	this.mux.Lock()
	defer this.mux.Unlock()
	consumer, err := NewConsumer(this.kafkaBootstrapUrl, this.group, topic, this.debug, func(topic string, msg []byte) error {
		return listener(msg)
	}, func(err error, consumer *Consumer) {
		debug.PrintStack()
		log.Fatal(err)
	})
	if err != nil {
		return err
	}
	this.consumers = append(this.consumers, consumer)
	return nil
}

func NewConsumer(zk string, groupid string, topic string, debug bool, listener func(topic string, msg []byte) error, errorhandler func(err error, consumer *Consumer)) (consumer *Consumer, err error) {
	consumer = &Consumer{groupId: groupid, kafkaBootstrapUrl: zk, topic: topic, listener: listener, errorhandler: errorhandler, debug: debug}
	err = consumer.start()
	return
}

type Consumer struct {
	count             int
	kafkaBootstrapUrl string
	groupId           string
	topic             string
	ctx               context.Context
	cancel            context.CancelFunc
	listener          func(topic string, msg []byte) error
	errorhandler      func(err error, consumer *Consumer)
	mux               sync.Mutex
	debug             bool
}

func (this *Consumer) Stop() {
	this.cancel()
}

func (this *Consumer) start() error {
	log.Println("DEBUG: consume topic: \"" + this.topic + "\"")
	this.ctx, this.cancel = context.WithCancel(context.Background())
	broker, err := GetBroker(this.kafkaBootstrapUrl)
	if err != nil {
		log.Println("ERROR: unable to get broker list", err)
		return err
	}
	err = InitTopic(this.kafkaBootstrapUrl, this.topic)
	if err != nil {
		log.Println("ERROR: unable to create topic", err)
		return err
	}
	r := kafka.NewReader(kafka.ReaderConfig{
		CommitInterval: 0, //synchronous commits
		Brokers:        broker,
		GroupID:        this.groupId,
		Topic:          this.topic,
		MaxWait:        1 * time.Second,
		Logger:         log.New(ioutil.Discard, "", 0),
		ErrorLogger:    log.New(ioutil.Discard, "", 0),
	})
	go func() {
		for {
			select {
			case <-this.ctx.Done():
				log.Println("close kafka reader ", this.topic)
				return
			default:
				m, err := r.FetchMessage(this.ctx)
				if err == io.EOF || err == context.Canceled {
					log.Println("close consumer for topic ", this.topic)
					return
				}
				if err != nil {
					log.Println("ERROR: while consuming topic ", this.topic, err)
					this.errorhandler(err, this)
					return
				}
				if this.debug {
					log.Println("consume: ", m.Time, m.Topic, string(m.Value))
				}
				err = this.listener(m.Topic, m.Value)
				if err != nil {
					log.Println("ERROR: unable to handle message (no commit)", err)
				} else {
					err = r.CommitMessages(this.ctx, m)
					if err != nil {
						log.Println("ERROR: unable to commit", err)
					}

				}
			}
		}
	}()
	return err
}

func (this *Consumer) Restart() {
	this.Stop()
	this.start()
}
