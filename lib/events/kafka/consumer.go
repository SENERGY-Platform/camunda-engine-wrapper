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
	"errors"
	"github.com/segmentio/kafka-go"
	"io"
	"log"
	"os"
	"runtime/debug"
	"sync"
	"time"
)

func (this *Kafka) Consume(topic string, listener func(delivery []byte) error) (err error) {
	this.mux.Lock()
	defer this.mux.Unlock()
	consumer, err := NewConsumer(this.broker, this.group, topic, this.debug, func(topic string, msg []byte) error {
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

func NewConsumer(broker string, groupid string, topic string, debug bool, listener func(topic string, msg []byte) error, errorhandler func(err error, consumer *Consumer)) (consumer *Consumer, err error) {
	consumer = &Consumer{groupId: groupid, broker: broker, topic: topic, listener: listener, errorhandler: errorhandler, debug: debug}
	err = consumer.start()
	return
}

type Consumer struct {
	count        int
	broker       string
	groupId      string
	topic        string
	ctx          context.Context
	cancel       context.CancelFunc
	listener     func(topic string, msg []byte) error
	errorhandler func(err error, consumer *Consumer)
	mux          sync.Mutex
	debug        bool
}

func (this *Consumer) Stop() {
	this.cancel()
}

func (this *Consumer) start() error {
	log.Println("DEBUG: consume topic: \"" + this.topic + "\"")
	this.ctx, this.cancel = context.WithCancel(context.Background())
	err := InitTopic(this.broker, this.topic)
	if err != nil {
		log.Println("ERROR: unable to create topic", err)
		return err
	}
	r := kafka.NewReader(kafka.ReaderConfig{
		CommitInterval:         0, //synchronous commits
		Brokers:                []string{this.broker},
		GroupID:                this.groupId,
		Topic:                  this.topic,
		MaxWait:                1 * time.Second,
		Logger:                 log.New(io.Discard, "", 0),
		ErrorLogger:            log.New(os.Stdout, "[KAFKA-ERR] ", log.LstdFlags),
		WatchPartitionChanges:  true,
		PartitionWatchInterval: time.Minute,
	})
	go func() {
		defer r.Close()
		defer log.Println("close consumer for topic ", this.topic)
		for {
			select {
			case <-this.ctx.Done():
				return
			default:
				m, err := r.FetchMessage(this.ctx)
				if err == io.EOF || err == context.Canceled {
					return
				}
				if err != nil {
					log.Println("ERROR: while consuming topic ", this.topic, err)
					this.errorhandler(err, this)
					return
				}

				err = retry(func() error {
					return this.listener(m.Topic, m.Value)
				}, func(n int64) time.Duration {
					return time.Duration(n) * time.Second
				}, 10*time.Minute)

				if err != nil {
					log.Println("ERROR: unable to handle message (no commit)", err)
					this.errorhandler(err, this)
				} else {
					err = r.CommitMessages(this.ctx, m)
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

func retry(f func() error, waitProvider func(n int64) time.Duration, timeout time.Duration) (err error) {
	err = errors.New("")
	start := time.Now()
	for i := int64(1); err != nil && time.Since(start) < timeout; i++ {
		err = f()
		if err != nil {
			log.Println("ERROR: kafka listener error:", err)
			wait := waitProvider(i)
			if time.Since(start)+wait < timeout {
				log.Println("ERROR: retry after:", wait.String())
				time.Sleep(wait)
			} else {
				return err
			}
		}
	}
	return err
}
