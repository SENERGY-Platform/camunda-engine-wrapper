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
	"io/ioutil"
	"net/url"

	"errors"
	"net/http"

	"encoding/json"

	"log"

	"github.com/SmartEnergyPlatform/util/http/request"
)

type StartResult struct {
	Id    string `json:"id"`
	Ended bool   `json:"ended"`
}

type OwnerWrapper struct {
	Owner string `json:"tenantId"`
}

type ProcessVariable struct {
	Name  string      `json:"name"`
	Value interface{} `json:"value"`
}

func startProcess(processDefinitionId string) (result []ProcessVariable, err error) {
	startResult := StartResult{}
	var code int
	err, _, code = request.Post(Config.ProcessEngineUrl+"/engine-rest/process-definition/"+url.QueryEscape(processDefinitionId)+"/start", map[string]string{}, &startResult)
	if err != nil {
		return
	}
	if code != http.StatusOK {
		err = errors.New("error on process start (status != 200)")
		return
	}
	result, err = getProcessVariables(startResult.Id, startResult.Ended)
	return
}

func startProcessGetId(processDefinitionId string) (result StartResult, err error) {
	var code int
	err, _, code = request.Post(Config.ProcessEngineUrl+"/engine-rest/process-definition/"+url.QueryEscape(processDefinitionId)+"/start", map[string]string{}, &result)
	if err == nil && code != http.StatusOK {
		err = errors.New("error on process start (status != 200)")
		return
	}
	return
}

func getProcessVariables(processDefinitionId string, processEnded bool) (result []ProcessVariable, err error) {
	if processEnded {
		err = request.Get(Config.ProcessEngineUrl+"/engine-rest/history/variable-instance?processInstanceId="+url.QueryEscape(processDefinitionId), &result)
	} else {
		err = request.Get(Config.ProcessEngineUrl+"/engine-rest/variable-instance?processInstanceId="+url.QueryEscape(processDefinitionId), &result)
	}
	return
}

func checkProcessDefinitionAccess(id string, userId string) (err error) {
	wrapper := OwnerWrapper{}
	err = request.Get(Config.ProcessEngineUrl+"/engine-rest/process-definition/"+url.QueryEscape(id), &wrapper)
	if err == nil && wrapper.Owner != userId {
		err = errors.New("access denied")
	}
	return
}

func checkDeploymentAccess(id string, userId string) (err error) {
	wrapper := OwnerWrapper{}
	err = request.Get(Config.ProcessEngineUrl+"/engine-rest/deployment/"+url.QueryEscape(id), &wrapper)
	if err != nil {
		log.Println("ERROR in request: ", Config.ProcessEngineUrl+"/engine-rest/deployment/"+url.QueryEscape(id), err)
	}
	if err == nil && wrapper.Owner != userId {
		err = errors.New("access denied")
	}
	return
}

func checkProcessInstanceAccess(id string, userId string) (err error) {
	wrapper := OwnerWrapper{}
	err = request.Get(Config.ProcessEngineUrl+"/engine-rest/process-instance/"+url.QueryEscape(id), &wrapper)
	if err == nil && wrapper.Owner != userId {
		err = errors.New("access denied")
	}
	return
}

func checkHistoryAccess(id string, userId string) (err error) {
	wrapper := OwnerWrapper{}
	err = request.Get(Config.ProcessEngineUrl+"/engine-rest/history/process-instance/"+url.QueryEscape(id), &wrapper)
	if err == nil && wrapper.Owner != userId {
		err = errors.New("access denied")
	}
	return
}

func getProcessInstanceIncidents(id string) (result interface{}, err error) {
	//"/engine-rest/incident?processInstanceId=" + processInstanceId
	err = request.Get(Config.ProcessEngineUrl+"/engine-rest/incident?processInstanceId="+url.QueryEscape(id), &result)
	return
}

func removeProcessInstance(id string) (err error) {
	////DELETE "/engine-rest/process-instance/" + processInstanceId
	client := &http.Client{}
	request, err := http.NewRequest("DELETE", Config.ProcessEngineUrl+"/engine-rest/process-instance/"+url.QueryEscape(id), nil)
	if err != nil {
		return
	}
	resp, err := client.Do(request)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if !(resp.StatusCode == 200 || resp.StatusCode == 204) {
		msg, _ := ioutil.ReadAll(resp.Body)
		err = errors.New("error on delete in engine for " + Config.ProcessEngineUrl+"/engine-rest/process-instance/"+url.QueryEscape(id) + ": " + resp.Status + " " + string(msg))
	}
	return
}

func removeProcessInstanceHistory(id string) (result interface{}, err error) {
	//DELETE "/engine-rest/history/process-instance/" + processInstanceId
	client := &http.Client{}
	request, err := http.NewRequest("DELETE", Config.ProcessEngineUrl+"/engine-rest/history/process-instance/"+url.QueryEscape(id), nil)
	if err != nil {
		return
	}
	resp, err := client.Do(request)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	err = json.NewDecoder(resp.Body).Decode(&result)
	return
}

func getProcessInstanceHistoricVariables(id string) (result interface{}, err error) {
	//"/engine-rest/history/variable-instance?processInstanceId=" + processInstanceId
	err = request.Get(Config.ProcessEngineUrl+"/engine-rest/history/variable-instance?processInstanceId="+url.QueryEscape(id), &result)
	return
}

func getProcessInstanceHistoryByProcessDefinition(id string) (result interface{}, err error) {
	//"/engine-rest/history/process-instance?processDefinitionId="
	err = request.Get(Config.ProcessEngineUrl+"/engine-rest/history/process-instance?processDefinitionId="+url.QueryEscape(id), &result)
	return
}
func getProcessInstanceHistoryByProcessDefinitionFinished(id string) (result interface{}, err error) {
	//"/engine-rest/history/process-instance?processDefinitionId="
	err = request.Get(Config.ProcessEngineUrl+"/engine-rest/history/process-instance?processDefinitionId="+url.QueryEscape(id)+"&finished=true", &result)
	return
}
func getProcessInstanceHistoryByProcessDefinitionUnfinished(id string) (result interface{}, err error) {
	//"/engine-rest/history/process-instance?processDefinitionId="
	err = request.Get(Config.ProcessEngineUrl+"/engine-rest/history/process-instance?processDefinitionId="+url.QueryEscape(id)+"&unfinished=true", &result)
	return
}

func getProcessInstanceHistoryList(userId string) (result interface{}, err error) {
	//"/engine-rest/process-instance"
	err = request.Get(Config.ProcessEngineUrl+"/engine-rest/history/process-instance?tenantIdIn="+url.QueryEscape(userId), &result)
	return
}
func getProcessInstanceHistoryListFinished(userId string) (result interface{}, err error) {
	//"/engine-rest/process-instance"
	err = request.Get(Config.ProcessEngineUrl+"/engine-rest/history/process-instance?tenantIdIn="+url.QueryEscape(userId)+"&finished=true", &result)
	return
}
func getProcessInstanceHistoryListUnfinished(userId string) (result interface{}, err error) {
	//"/engine-rest/process-instance"
	err = request.Get(Config.ProcessEngineUrl+"/engine-rest/history/process-instance?tenantIdIn="+url.QueryEscape(userId)+"&unfinished=true", &result)
	return
}
func getProcessInstanceCount(userId string) (result interface{}, err error) {
	//"/engine-rest/process-instance/count"
	err = request.Get(Config.ProcessEngineUrl+"/engine-rest/process-instance/count?tenantIdIn="+url.QueryEscape(userId), &result)
	return
}
func getProcessInstanceList(userId string) (result interface{}, err error) {
	//"/engine-rest/process-instance"
	err = request.Get(Config.ProcessEngineUrl+"/engine-rest/process-instance?tenantIdIn="+url.QueryEscape(userId), &result)
	return
}
func getProcessDefinitionIncident(definitionId string) (result interface{}, err error) {
	//"/engine-rest/incident/count?processDefinitionId=" + deployment.processDefintionId
	err = request.Get(Config.ProcessEngineUrl+"/engine-rest/incident/count?processDefinitionId="+url.QueryEscape(definitionId), &result)
	return
}
func getProcessDefinition(id string) (result interface{}, err error) {
	//"/engine-rest/process-definition/" + processDefinitionId
	err = request.Get(Config.ProcessEngineUrl+"/engine-rest/process-definition/"+url.QueryEscape(id), &result)
	return
}
func getProcessDefinitionDiagram(id string) (resp *http.Response, err error) {
	// "/engine-rest/process-definition/" + processDefinitionId + "/diagram"
	resp, err = http.Get(Config.ProcessEngineUrl + "/engine-rest/process-definition/" + url.QueryEscape(id) + "/diagram")
	return
}
func getDeploymentList(userId string, params url.Values) (result interface{}, err error) {
	// "/engine-rest/deployment?tenantIdIn="+userId
	params.Del("tenantIdIn")
	path := Config.ProcessEngineUrl+"/engine-rest/deployment?tenantIdIn="+url.QueryEscape(userId)+"&"+params.Encode()
	err = request.Get(path, &result)
	return
}
func getDefinitionByDeployment(id string) (result interface{}, err error) {
	//"/engine-rest/process-definition?deploymentId=
	err = request.Get(Config.ProcessEngineUrl+"/engine-rest/process-definition?deploymentId="+url.QueryEscape(id), &result)
	return
}
func getDeployment(deploymentId string) (result interface{}, err error) {
	//"/engine-rest/deployment/" + id
	err = request.Get(Config.ProcessEngineUrl+"/engine-rest/deployment/"+url.QueryEscape(deploymentId), &result)
	return
}
