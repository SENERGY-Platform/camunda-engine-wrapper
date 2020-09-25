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

package lib

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/etree"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/kafka"
	"log"
	"strconv"
	"strings"
	"time"
)

var cqrs kafka.Interface

type DeploymentV1 struct {
	Id   string `json:"id"`
	Xml  string `json:"xml"`
	Svg  string `json:"svg"`
	Name string `json:"name"`
}

type DeploymentV2 struct {
	Id      string  `json:"id"`
	Name    string  `json:"name"`
	Diagram Diagram `json:"diagram"`
}

type Diagram struct {
	XmlDeployed string `json:"xml_deployed"`
	Svg         string `json:"svg"`
}

type DeploymentCommand struct {
	Command      string        `json:"command"`
	Id           string        `json:"id"`
	Owner        string        `json:"owner"`
	Deployment   *DeploymentV1 `json:"deployment"`
	DeploymentV2 *DeploymentV2 `json:"deployment_v2"`
	Source       string        `json:"source,omitempty"`
}

type KafkaIncidentsCommand struct {
	Command             string `json:"command"`
	MsgVersion          int64  `json:"msg_version"`
	ProcessDefinitionId string `json:"process_definition_id,omitempty"`
	ProcessInstanceId   string `json:"process_instance_id,omitempty"`
}

func InitEventSourcing(kafka kafka.Interface) (err error) {
	cqrs = kafka
	err = cqrs.Consume(Config.DeploymentTopic, func(delivery []byte) error {
		command := DeploymentCommand{}
		err = json.Unmarshal(delivery, &command)
		if err != nil {
			log.Println("ERROR: unable to parse cqrs event as json \n", err, "\n --> ignore event \n", string(delivery))
			return nil
		}
		log.Println("cqrs receive ", string(delivery))
		switch command.Command {
		case "PUT":
			owner, id, name, xml, svg, source, err := parsePutCommand(command)
			if err != nil {
				return err
			}
			return handleDeploymentCreate(owner, id, name, xml, svg, source)
		case "POST":
			log.Println("WARNING: deprecated event type POST")
			return nil
		case "DELETE":
			return handleDeploymentDelete(command.Id, command.Owner)
		default:
			log.Println("WARNING: unknown event type", string(delivery))
			return nil
		}
	})
	return err
}

func parsePutCommand(command DeploymentCommand) (owner string, id string, name string, xml string, svg string, source string, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("recovered error %v", r)
		}
	}()
	if command.Deployment != nil {
		return parseV1Command(command)
	} else if command.DeploymentV2 != nil {
		return parseV2Command(command)
	}
	err = errors.New("unknown version")
	return
}

func parseV1Command(command DeploymentCommand) (owner string, id string, name string, xml string, svg string, source string, err error) {
	return command.Owner, command.Id, command.Deployment.Name, command.Deployment.Xml, command.Deployment.Svg, command.Source, nil
}

func parseV2Command(command DeploymentCommand) (owner string, id string, name string, xml string, svg string, source string, err error) {
	return command.Owner, command.Id, command.DeploymentV2.Name, command.DeploymentV2.Diagram.XmlDeployed, command.DeploymentV2.Diagram.Svg, command.Source, err
}

func handleDeploymentDelete(vid string, userId string) error {
	id, exists, err := getDeploymentId(vid)
	if err != nil {
		return err
	}
	if !exists {
		return nil
	}

	err = deleteIncidentsByDeploymentId(id, userId)
	if err != nil {
		return err
	}

	commit, rollback, err := removeVidRelation(vid, id)
	if err != nil {
		return err
	}
	if userId != "" {
		err = RemoveProcess(id, userId)
	} else {
		err = RemoveProcessFromAllShards(id)
	}
	err = RemoveProcess(id, userId)
	if err != nil {
		rollback()
	} else {
		commit()
	}
	return err
}

func deleteIncidentsByDeploymentId(id string, userId string) (err error) {
	definitions, err := getRawDefinitionsByDeployment(id, userId)
	if err != nil {
		return err
	}
	for _, definition := range definitions {
		err = PublishIncidentsDeleteByProcessDefinitionEvent(definition.Id)
		if err != nil {
			return err
		}
	}
	return nil
}

func PublishIncidentsDeleteByProcessDefinitionEvent(definitionId string) error {
	command := KafkaIncidentsCommand{
		Command:             "DELETE",
		ProcessDefinitionId: definitionId,
		MsgVersion:          3,
	}
	payload, err := json.Marshal(command)
	if err != nil {
		return err
	}
	return cqrs.Publish(Config.IncidentTopic, definitionId, payload)
}

func PublishIncidentDeleteByProcessInstanceEvent(instanceId string, definitionId string) error {
	command := KafkaIncidentsCommand{
		Command:           "DELETE",
		ProcessInstanceId: instanceId,
		MsgVersion:        3,
	}
	payload, err := json.Marshal(command)
	if err != nil {
		return err
	}
	return cqrs.Publish(Config.IncidentTopic, definitionId, payload)
}

func handleDeploymentCreate(owner string, id string, name string, xml string, svg string, source string) (err error) {
	err = cleanupExistingDeployment(id, owner)
	if err != nil {
		return err
	}
	if !validateXml(xml) {
		log.Println("ERROR: got invalid xml, replace with default")
		xml = createBlankProcess()
		svg = createBlankSvg()
	}
	if Config.Debug {
		log.Println("deploy process", id, name, xml)
	}
	deploymentId, err := DeployProcess(name, xml, svg, owner, source)
	if err != nil {
		log.Println("WARNING: unable to deploy process to camunda ", err)
		return err
	}
	if Config.Debug {
		log.Println("save vid relation", id, deploymentId)
	}
	err = saveVidRelation(id, deploymentId)
	if err != nil {
		log.Println("WARNING: unable to publish deployment saga \n", err, "\nremove deployed process")
		removeErr := RemoveProcess(deploymentId, owner)
		if removeErr != nil {
			log.Println("ERROR: unable to remove deployed process", deploymentId, err)
		}
		return err
	}
	return err
}

func createBlankSvg() string {
	return `<svg xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" version="1.2" id="Layer_1" x="0px" y="0px" viewBox="0 0 20 16" xml:space="preserve">
<path fill="#D61F33" d="M10,0L0,16h20L10,0z M11,13.908H9v-2h2V13.908z M9,10.908v-6h2v6H9z"/>
</svg>`
}

func createBlankProcess() string {
	templ := `<bpmn:definitions xmlns:xsi='http://www.w3.org/2001/XMLSchema-instance' xmlns:bpmn='http://www.omg.org/spec/BPMN/20100524/MODEL' xmlns:bpmndi='http://www.omg.org/spec/BPMN/20100524/DI' xmlns:dc='http://www.omg.org/spec/DD/20100524/DC' id='Definitions_1' targetNamespace='http://bpmn.io/schema/bpmn'><bpmn:process id='PROCESSID' isExecutable='true'><bpmn:startEvent id='StartEvent_1'/></bpmn:process><bpmndi:BPMNDiagram id='BPMNDiagram_1'><bpmndi:BPMNPlane id='BPMNPlane_1' bpmnElement='PROCESSID'><bpmndi:BPMNShape id='_BPMNShape_StartEvent_2' bpmnElement='StartEvent_1'><dc:Bounds x='173' y='102' width='36' height='36'/></bpmndi:BPMNShape></bpmndi:BPMNPlane></bpmndi:BPMNDiagram></bpmn:definitions>`
	return strings.Replace(templ, "PROCESSID", "id_"+strconv.FormatInt(time.Now().Unix(), 10), 1)
}

func validateXml(xmlStr string) bool {
	if xmlStr == "" {
		return false
	}
	err := etree.NewDocument().ReadFromString(xmlStr)
	if err != nil {
		log.Println("ERROR: unable to parse xml", err)
		return false
	}
	err = xml.Unmarshal([]byte(xmlStr), new(interface{}))
	if err != nil {
		log.Println("ERROR: unable to parse xml", err)
		return false
	}
	return true
}

func cleanupExistingDeployment(vid string, userId string) error {
	exists, err := vidExists(vid)
	if err != nil {
		return err
	}
	if exists {
		return handleDeploymentDelete(vid, userId)
	}
	return nil
}

func CloseEventSourcing() {
	cqrs.Close()
}
