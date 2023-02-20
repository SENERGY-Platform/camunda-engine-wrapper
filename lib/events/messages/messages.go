/*
 * Copyright 2023 InfAI (CC SES)
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

package messages

import "time"

const CurrentVersion int64 = 3

type Deployment struct {
	Version          int64             `json:"version"`
	Id               string            `json:"id"`
	Name             string            `json:"name"`
	Diagram          Diagram           `json:"diagram"`
	IncidentHandling *IncidentHandling `json:"incident_handling,omitempty"`
}

type IncidentHandling struct {
	Restart              bool `json:"restart"`
	Notify               bool `json:"notify"`
	RestartIsValidOption bool `json:"restart_is_valid_option"`
}

type Diagram struct {
	XmlDeployed string `json:"xml_deployed"`
	Svg         string `json:"svg"`
}

type DeploymentCommand struct {
	Command    string      `json:"command"`
	Id         string      `json:"id"`
	Owner      string      `json:"owner"`
	Deployment *Deployment `json:"deployment"`
	Source     string      `json:"source,omitempty"`
	Version    int64       `json:"version"`
}

type VersionWrapper struct {
	Command string `json:"command"`
	Id      string `json:"id"`
	Version int64  `json:"version"`
	Owner   string `json:"owner"`
}

type KafkaIncidentsCommand struct {
	Command             string      `json:"command"`
	MsgVersion          int64       `json:"msg_version"`
	Incident            *Incident   `json:"incident,omitempty"`
	Handler             *OnIncident `json:"handler,omitempty"`
	ProcessDefinitionId string      `json:"process_definition_id,omitempty"`
	ProcessInstanceId   string      `json:"process_instance_id,omitempty"`
}

type Incident struct {
	Id                  string    `json:"id" bson:"id"`
	MsgVersion          int64     `json:"msg_version,omitempty" bson:"msg_version,omitempty"` //from version 3 onward will be set in KafkaIncidentsCommand and be copied to this field
	ExternalTaskId      string    `json:"external_task_id" bson:"external_task_id"`
	ProcessInstanceId   string    `json:"process_instance_id" bson:"process_instance_id"`
	ProcessDefinitionId string    `json:"process_definition_id" bson:"process_definition_id"`
	WorkerId            string    `json:"worker_id" bson:"worker_id"`
	ErrorMessage        string    `json:"error_message" bson:"error_message"`
	Time                time.Time `json:"time" bson:"time"`
	TenantId            string    `json:"tenant_id" bson:"tenant_id"`
	DeploymentName      string    `json:"deployment_name" bson:"deployment_name"`
}

type OnIncident struct {
	ProcessDefinitionId string `json:"process_definition_id" bson:"process_definition_id"`
	Restart             bool   `json:"restart" bson:"restart"`
	Notify              bool   `json:"notify" bson:"notify"`
}

type DoneNotification struct {
	Command string `json:"command"`
	Id      string `json:"id"`
	Handler string `json:"handler"`
}
