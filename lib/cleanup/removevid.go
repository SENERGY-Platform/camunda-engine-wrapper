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

package cleanup

import (
	"encoding/json"
	"fmt"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/configuration"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/events/kafka"
)

func RemoveVid(config configuration.Config, ids []string) error {
	cqrs, err := kafka.Init(config.KafkaUrl, config.KafkaGroup, config.Debug)
	if err != nil {
		return err
	}
	return RemoveVidWithCqrs(cqrs, config, ids)
}

func RemoveVidWithCqrs(cqrs kafka.Interface, config configuration.Config, ids []string) error {
	for _, id := range ids {
		err := removeVidByEvent(cqrs, config.DeploymentTopic, id)
		if err != nil {
			return err
		}
	}
	return nil
}

func removeVidByEvent(cqrs kafka.Interface, topic string, vid string) error {
	fmt.Println("remove vid", vid)
	command := DeploymentDeleteCommand{Id: vid, Command: "DELETE"}
	payload, err := json.Marshal(command)
	if err != nil {
		return err
	}
	return cqrs.Publish(topic, vid, payload)
}

type DeploymentDeleteCommand struct {
	Command string `json:"command"`
	Id      string `json:"id"`
	Owner   string `json:"owner"`
	Source  string `json:"source,omitempty"`
}
