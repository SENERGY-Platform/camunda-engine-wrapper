/*
 * Copyright 2025 InfAI (CC SES)
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

package controller

import (
	"encoding/xml"
	"errors"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/etree"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/camunda"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/configuration"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/model"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/processio"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/vid"
	"github.com/SENERGY-Platform/process-incident-api/lib/client"
	"log"
	"net/http"
)

type Controller struct {
	config    configuration.Config
	camunda   *camunda.Camunda
	vid       *vid.Vid
	processIo *processio.ProcessIo
}

func New(config configuration.Config, camunda *camunda.Camunda, vid *vid.Vid, processIo *processio.ProcessIo) *Controller {
	return &Controller{
		config:    config,
		camunda:   camunda,
		vid:       vid,
		processIo: processIo,
	}
}

func (this *Controller) Deploy(depl model.DeploymentMessage) (err error, code int) {
	xml, err := SecureProcessScripts(depl.Diagram.XmlDeployed)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	if depl.Id == "" {
		return errors.New("no deployment id provided"), http.StatusBadRequest
	}
	if depl.UserId == "" {
		return errors.New("no user id provided"), http.StatusBadRequest
	}
	if depl.Diagram.Svg == "" {
		return errors.New("no svg provided"), http.StatusBadRequest
	}
	if !validateXml(xml) {
		return errors.New("invalid bpmn"), http.StatusBadRequest
	}
	err = this.cleanupExistingDeployment(depl.UserId, depl.Id)
	if err != nil {
		return err, http.StatusInternalServerError
	}

	if this.config.Debug {
		log.Println("deploy process", depl.Id, depl.Name, xml)
	}
	deploymentId, err := this.camunda.DeployProcess(depl.Name, xml, depl.Diagram.Svg, depl.UserId, depl.Source)
	if err != nil {
		log.Println("WARNING: unable to deploy process to camunda ", err)
		return err, http.StatusInternalServerError
	}

	if depl.IncidentHandling != nil {
		definitions, err := this.camunda.GetRawDefinitionsByDeployment(deploymentId, depl.UserId)
		if err != nil {
			removeErr := this.camunda.RemoveProcess(deploymentId, depl.UserId)
			if removeErr != nil {
				log.Println("ERROR: unable to remove deployed process", deploymentId, removeErr, err)
			}
			return err, http.StatusInternalServerError
		}
		if len(definitions) == 0 {
			log.Println("WARNING: no definitions for deployment found --> no incident handling deployed")
		}
		for _, definition := range definitions {
			err, _ = client.New(this.config.IncidentApiUrl).SetOnIncidentHandler(client.InternalAdminToken, client.OnIncident{
				ProcessDefinitionId: definition.Id,
				Restart:             depl.IncidentHandling.Restart,
				Notify:              depl.IncidentHandling.Notify,
			})
			if err != nil {
				removeErr := this.camunda.RemoveProcess(deploymentId, depl.UserId)
				if removeErr != nil {
					log.Println("ERROR: unable to remove deployed process", deploymentId, removeErr, err)
				}
				return err, http.StatusInternalServerError
			}
		}
	}
	if this.config.Debug {
		log.Println("save vid relation", depl.Id, deploymentId)
	}
	err = this.vid.SaveVidRelation(depl.Id, deploymentId)
	if err != nil {
		log.Println("WARNING: unable to publish deployment saga \n", err, "\nremove deployed process")
		removeErr := this.camunda.RemoveProcess(deploymentId, depl.UserId)
		if removeErr != nil {
			log.Println("ERROR: unable to remove deployed process", deploymentId, removeErr, err)
		}
		return err, http.StatusInternalServerError
	}
	return err, http.StatusOK
}

func (this *Controller) DeleteDeployment(userId string, vid string) error {
	id, exists, err := this.vid.GetDeploymentId(vid)
	if err != nil {
		return err
	}
	if !exists {
		return nil
	}

	err = this.deleteIncidentsByDeploymentId(id, userId)
	if err != nil {
		return err
	}

	err = this.deleteIoVariablesByDeploymentId(id, userId)
	if err != nil {
		return err
	}

	commit, rollback, err := this.vid.RemoveVidRelation(vid, id)
	if err != nil {
		return err
	}
	if userId != "" {
		err = this.camunda.RemoveProcess(id, userId)
	} else {
		err = this.camunda.RemoveProcessFromAllShards(id)
	}
	if err != nil {
		_ = rollback()
	} else {
		return commit()
	}
	return err
}

func (this *Controller) DeleteHistoricProcessInstance(userId string, instanceId string) (err error, code int) {
	_, err = this.camunda.CheckHistoryAccess(instanceId, userId)
	if err != nil {
		return errors.New("access denied"), http.StatusUnauthorized
	}
	err, code = client.New(this.config.IncidentApiUrl).DeleteIncidentByProcessInstanceId(client.InternalAdminToken, instanceId)
	if err != nil {
		return err, code
	}
	err = this.camunda.RemoveProcessInstanceHistory(instanceId, userId)
	if err != nil {
		return err, code
	}
	return nil, http.StatusOK
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

func (this *Controller) cleanupExistingDeployment(userId string, vid string) error {
	exists, err := this.vid.VidExists(vid)
	if err != nil {
		return err
	}
	if exists {
		return this.DeleteDeployment(userId, vid)
	}
	return nil
}

func (this *Controller) deleteIncidentsByDeploymentId(id string, userId string) (err error) {
	definitions, err := this.camunda.GetRawDefinitionsByDeployment(id, userId)
	if err != nil {
		return err
	}
	for _, definition := range definitions {
		err, _ = client.New(this.config.IncidentApiUrl).DeleteIncidentByProcessDefinitionId(client.InternalAdminToken, definition.Id)
		if err != nil {
			return err
		}
	}
	return nil
}

func (this *Controller) deleteIoVariablesByDeploymentId(id string, userId string) (err error) {
	if this.processIo != nil {
		definitions, err := this.camunda.GetRawDefinitionsByDeployment(id, userId)
		if err != nil {
			return err
		}
		for _, definition := range definitions {
			err = this.processIo.DeleteProcessDefinition(definition.Id)
			if err != nil {
				return err
			}
		}
		return nil
	}
	return nil
}
