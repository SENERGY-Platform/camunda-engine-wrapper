/*
 * Copyright 2018 InfAI (CC SES)
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

package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/configuration"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/events/kafka"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/shardmigration"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/vidcleanup"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	defer fmt.Println("exit application")
	configLocation := flag.String("config", "config.json", "configuration file")

	migrateShard := flag.String("migrate_shard", "", "if set this program will only migrate the users of the given camunda-engine to the shard database")

	vidCleanup := flag.Bool("vid_cleanup", false, "if true, the program will only clean the vid-database and camunda from inconsistencies")

	flag.Parse()

	config, err := configuration.LoadConfig(*configLocation)
	if err != nil {
		log.Fatal("unable to load config", err)
	}

	if *migrateShard != "" {
		err = shardmigration.Run(*migrateShard, config.ShardingDb, 100)
		if err != nil {
			log.Fatal("unable to do shard migration:", err)
		}
	} else if *vidCleanup {
		cqrs, err := kafka.Init(config.ZookeeperUrl, config.KafkaGroup, config.Debug)
		if err != nil {
			log.Fatal("unable to connect to kafka for do vid cleanup:", err)
		}
		err = vidcleanup.ClearUnlinkedDeployments(config.WrapperDb, config.DeploymentTopic, cqrs, 24*time.Hour)
		if err != nil {
			log.Fatal("unable to do vid cleanup:", err)
		}
	} else {
		wrapper(config)
	}
}

func wrapper(config configuration.Config) {
	ctx, cancel := context.WithCancel(context.Background())
	err := lib.Wrapper(ctx, config)
	if err != nil {
		log.Fatalf("FATAL: %+v", err)
	}
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
	sig := <-shutdown
	log.Println("received shutdown signal", sig)
	cancel()
}
