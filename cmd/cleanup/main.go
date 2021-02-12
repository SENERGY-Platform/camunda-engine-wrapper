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
	"bufio"
	"errors"
	"flag"
	"fmt"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/cleanup"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/configuration"
	"log"
	"os"
	"strings"
	"time"
)

const FindUnlinkedVidArg = "unlinked-vid"
const FindUnlinkedPidArg = "unlinked-pid"
const RemoveVidArg = "remove-vid"
const RemovePidArg = "remove-pid"

func main() {
	configLocation := flag.String("config", "config.json", "configuration file")
	noConfirmation := flag.Bool("no-confirmation", false, "dont confirm remove operation")
	flag.Parse()

	args := flag.Args()
	if len(args) == 0 {
		log.Fatal("missing args")
	}

	configuration.LogEnvConfig = false
	config, err := configuration.LoadConfig(*configLocation)
	if err != nil {
		log.Fatal("unable to load config", err)
	}

	switch args[0] {
	case FindUnlinkedVidArg:
		err = printUnlinkedVid(config)
	case FindUnlinkedPidArg:
		err = printUnlinkedPid(config)
	case RemoveVidArg:
		err = removeVids(config, args[1:], *noConfirmation)
	case RemovePidArg:
		err = removePids(args[1:], *noConfirmation)
	default:
		err = errors.New(fmt.Sprint("unknown args'", args))
	}
	if err != nil {
		log.Fatal(err)
	}
}

func removePids(inputs []string, noConfirmation bool) error {
	if len(inputs)%2 != 0 {
		return errors.New("expect even count of arguments, with pairs of shard-urls and pid")
	}
	shardToPids := map[string][]string{}
	for i := 0; i < len(inputs)-1; i += 2 {
		shardUrl := inputs[i]
		shardToPids[shardUrl] = append(shardToPids[shardUrl], inputs[i+1])
	}

	fmt.Println("this will delete", len(inputs)/2, "pids in", len(shardToPids), "shards")

	if !noConfirmation {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("continue? (y/n): ")
		in, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		if strings.TrimSpace(in) != "y" && strings.TrimSpace(in) != "Y" {
			return nil
		}
	}

	return cleanup.RemovePid(shardToPids)
}

func removeVids(config configuration.Config, inputs []string, noConfirmation bool) error {
	fmt.Println("this will delete", len(inputs), "vids")

	if !noConfirmation {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("continue? (y/n): ")
		in, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		if strings.TrimSpace(in) != "y" && strings.TrimSpace(in) != "Y" {
			return nil
		}
	}

	return cleanup.RemoveVid(config, inputs)
}

func printUnlinkedVid(config configuration.Config) error {
	unlinkedVid, err := cleanup.FindUnlinkedVid(config)
	if err != nil {
		return err
	}
	for _, pid := range unlinkedVid {
		fmt.Println(pid)
	}
	return nil
}

func printUnlinkedPid(config configuration.Config) error {
	unlinkedPid, err := cleanup.FindUnlinkedPid(config, 24*time.Hour)
	if err != nil {
		return err
	}
	for shard, pids := range unlinkedPid {
		for _, pid := range pids {
			fmt.Println(shard, pid)
		}
	}
	return nil
}
