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
	"encoding/json"
	"github.com/SmartEnergyPlatform/amqp-wrapper-lib"
	"log"
)

var amqp *amqp_wrapper_lib.Connection

type AbstractProcess struct {
	Xml           string         `json:"xml"`
	Name          string         `json:"name"`
	AbstractTasks interface{} `json:"abstract_tasks"`
	AbstractDataExportTasks interface{} `json:"abstract_data_export_tasks"`
	ReceiveTasks  interface{}     `json:"receive_tasks"`
	MsgEvents     interface{}     `json:"msg_events"`
	TimeEvents    interface{}    `json:"time_events"`
}

type DeploymentRequest struct {
	Svg     string          `json:"svg"`
	Process AbstractProcess `json:"process"`
}

type DeploymentCommand struct {
	Command    string          		 	`json:"command"`
	Id         string           		`json:"id"`
	Owner      string           		`json:"owner"`
	DeploymentXml string				`json:"deployment_xml"`
	Deployment DeploymentRequest	`json:"deployment"`
}

func InitEventSourcing()(err error){
	amqp, err = amqp_wrapper_lib.Init(Config.AmqpUrl, []string{Config.AmqpDeploymentTopic}, Config.AmqpReconnectTimeout)
	if err != nil {
		return err
	}
	err = amqp.Consume(Config.AmqpConsumerName + "_" +Config.AmqpDeploymentTopic, Config.AmqpDeploymentTopic, func(delivery []byte) error {
		command := DeploymentCommand{}
		err = json.Unmarshal(delivery, &command)
		if err != nil {
			log.Println("ERROR: unable to parse amqp event as json \n", err, "\n --> ignore event \n", string(delivery))
			return nil
		}
		log.Println("amqp receive ", string(delivery))
		switch command.Command {
		case "PUT":
			return nil
		case "POST":
			return handleDeploymentCreate(command)
		case "DELETE":
			return handleDeploymentDelete(command)
		default:
			log.Println("WARNING: unknown event type", string(delivery))
			return nil
		}
	})
	return err
}

func handleDeploymentDelete(command DeploymentCommand) error {
	return RemoveProcess(command.Id)
}

func handleDeploymentCreate(command DeploymentCommand)error{
	processId, err := DeployProcess(command.Deployment.Process.Name, command.DeploymentXml, command.Deployment.Svg, command.Owner)
	if err != nil {
		log.Println("WARNING: unable to deploy process to camunda ", err)
		return err
	}
	err = PublishDeploymentSaga(processId, command)
	if err != nil {
		log.Println("WARNING: unable to publish deployment saga \n", err, "\nremove deployed process")
		removeErr := RemoveProcess(processId)
		if removeErr != nil {
			log.Println("ERROR: unable to remove deployed process", processId, err)
		}
		return err
	}
	return err
}

func CloseEventSourcing(){
	amqp.Close()
}

func PublishDeploymentSaga(id string, command DeploymentCommand)error{
	command.Id = id
	command.Command = "PUT"
	payload, err := json.Marshal(command)
	if err != nil {
		return err
	}
	return amqp.Publish(Config.AmqpDeploymentTopic, payload)
}

