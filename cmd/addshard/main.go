/*
 * Copyright 2021 InfAI (CC SES)
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
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/configuration"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/shardmigration"
	"log"
)

func main() {
	configLocation := flag.String("config", "config.json", "configuration file")

	flag.Parse()

	args := flag.Args()
	if len(args) != 1 {
		log.Fatal("expect the camunda-url as argument, got", args)
	}

	config, err := configuration.LoadConfig(*configLocation)
	if err != nil {
		log.Fatal("unable to load config", err)
	}

	err = shardmigration.Run(args[0], config.ShardingDb, 100)
	if err != nil {
		log.Fatal("unable to do shard migration:", err)
	}
}
