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
	"flag"
	"fmt"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/kafka"
	"log"
	"time"
)

func main() {
	defer fmt.Println("exit application")
	configLocation := flag.String("config", "config.json", "configuration file")
	flag.Parse()

	err := lib.LoadConfig(*configLocation)
	if err != nil {
		log.Fatal("unable to load config", err)
	}

	cqrs, err := kafka.Init(lib.Config.ZookeeperUrl, lib.Config.KafkaGroup, lib.Config.KafkaDebug)
	if err != nil {
		log.Fatal("unable to init kafka connection", err)
	}

	err = lib.InitEventSourcing(cqrs)
	if err != nil {
		log.Fatal("unable to start eventsourcing", err)
	}

	defer lib.CloseEventSourcing()

	if lib.Config.MaintenanceTime > 0 {
		log.Println("MAINTENANCE: ", lib.ClearUnlinkedDeployments())
		ticker := time.NewTicker(time.Duration(lib.Config.MaintenanceTime) * time.Hour)
		defer ticker.Stop()
		go func() {
			for tick := range ticker.C {
				log.Println("MAINTENANCE: ", tick, lib.ClearUnlinkedDeployments())
			}
		}()
	}

	lib.InitApi()
}
