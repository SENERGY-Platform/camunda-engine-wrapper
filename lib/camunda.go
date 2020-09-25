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
	"io/ioutil"
	"net/url"
	"strings"
	"time"

	"errors"
	"net/http"

	"encoding/json"

	"log"

	"github.com/SmartEnergyPlatform/util/http/request"
)

func startProcess(processDefinitionId string, userId string) (err error) {
	shard, err := GetUserShard(userId)
	if err != nil {
		return err
	}
	startResult := ProcessInstance{}
	var code int
	err, _, code = request.Post(shard+"/engine-rest/process-definition/"+url.QueryEscape(processDefinitionId)+"/start", map[string]string{}, &startResult)
	if err != nil {
		return
	}
	if code != http.StatusOK {
		err = errors.New("error on process start (status != 200)")
		return
	}
	return
}

func startProcessGetId(processDefinitionId string, userId string) (result ProcessInstance, err error) {
	shard, err := GetUserShard(userId)
	if err != nil {
		return result, err
	}
	var code int
	err, _, code = request.Post(shard+"/engine-rest/process-definition/"+url.QueryEscape(processDefinitionId)+"/start", map[string]string{}, &result)
	if err == nil && code != http.StatusOK {
		err = errors.New("error on process start (status != 200)")
		return
	}
	return
}

func checkProcessDefinitionAccess(id string, userId string) (err error) {
	shard, err := GetUserShard(userId)
	if err != nil {
		return err
	}
	definition := ProcessDefinition{}
	err = request.Get(shard+"/engine-rest/process-definition/"+url.QueryEscape(id), &definition)
	if err == nil && definition.TenantId != userId {
		err = errors.New("access denied")
	}
	return
}

func checkDeploymentAccess(vid string, userId string) (err error) {
	shard, err := GetUserShard(userId)
	if err != nil {
		return err
	}
	id, exists, err := getDeploymentId(vid)
	if err != nil {
		return err
	}
	if !exists {
		return UnknownVid
	}
	wrapper := Deployment{}
	err = request.Get(shard+"/engine-rest/deployment/"+url.QueryEscape(id), &wrapper)
	if err != nil {
		return err
	}
	if wrapper.Id == "" {
		return CamundaDeploymentUnknown
	}
	if wrapper.TenantId != userId {
		err = AccessDenied
	}
	return
}

func checkProcessInstanceAccess(id string, userId string) (err error) {
	shard, err := GetUserShard(userId)
	if err != nil {
		return err
	}
	wrapper := ProcessInstance{}
	err = request.Get(shard+"/engine-rest/process-instance/"+url.QueryEscape(id), &wrapper)
	if err == nil && wrapper.TenantId != userId {
		err = errors.New("access denied")
	}
	return
}

func checkHistoryAccess(id string, userId string) (definitionId string, err error) {
	shard, err := GetUserShard(userId)
	if err != nil {
		return definitionId, err
	}
	wrapper := HistoricProcessInstance{}
	err = request.Get(shard+"/engine-rest/history/process-instance/"+url.QueryEscape(id), &wrapper)
	if err == nil && wrapper.TenantId != userId {
		err = errors.New("access denied")
	}
	return wrapper.ProcessDefinitionId, err
}

func removeProcessInstance(id string, userId string) (err error) {
	////DELETE "/engine-rest/process-instance/" + processInstanceId
	shard, err := GetUserShard(userId)
	if err != nil {
		return err
	}
	client := &http.Client{}
	request, err := http.NewRequest("DELETE", shard+"/engine-rest/process-instance/"+url.QueryEscape(id)+"?skipIoMappings=true", nil)
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
		err = errors.New("error on delete in engine for " + shard + "/engine-rest/process-instance/" + url.QueryEscape(id) + ": " + resp.Status + " " + string(msg))
	}
	return
}

func removeProcessInstanceHistory(id string, userId string) (err error) {
	shard, err := GetUserShard(userId)
	if err != nil {
		return err
	}
	//DELETE "/engine-rest/history/process-instance/" + processInstanceId
	client := &http.Client{}
	request, err := http.NewRequest("DELETE", shard+"/engine-rest/history/process-instance/"+url.QueryEscape(id), nil)
	if err != nil {
		return
	}
	resp, err := client.Do(request)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if err == nil && !(resp.StatusCode == 200 || resp.StatusCode == 204) {
		msg, _ := ioutil.ReadAll(resp.Body)
		err = errors.New("error on delete in engine for " + shard + "/engine-rest/history/process-instance/" + url.QueryEscape(id) + ": " + resp.Status + " " + string(msg))
	}
	return
}

func getProcessInstanceHistoryByProcessDefinition(id string, userId string) (result HistoricProcessInstances, err error) {
	shard, err := GetUserShard(userId)
	if err != nil {
		return result, err
	}
	//"/engine-rest/history/process-instance?processDefinitionId="
	err = request.Get(shard+"/engine-rest/history/process-instance?processDefinitionId="+url.QueryEscape(id), &result)
	return
}
func getProcessInstanceHistoryByProcessDefinitionFinished(id string, userId string) (result HistoricProcessInstances, err error) {
	shard, err := GetUserShard(userId)
	if err != nil {
		return result, err
	}
	//"/engine-rest/history/process-instance?processDefinitionId="
	err = request.Get(shard+"/engine-rest/history/process-instance?processDefinitionId="+url.QueryEscape(id)+"&finished=true", &result)
	return
}
func getProcessInstanceHistoryByProcessDefinitionUnfinished(id string, userId string) (result HistoricProcessInstances, err error) {
	shard, err := GetUserShard(userId)
	if err != nil {
		return result, err
	}
	//"/engine-rest/history/process-instance?processDefinitionId="
	err = request.Get(shard+"/engine-rest/history/process-instance?processDefinitionId="+url.QueryEscape(id)+"&unfinished=true", &result)
	return
}

func getProcessInstanceHistoryList(userId string) (result HistoricProcessInstances, err error) {
	shard, err := GetUserShard(userId)
	if err != nil {
		return result, err
	}
	//"/engine-rest/process-instance"
	err = request.Get(shard+"/engine-rest/history/process-instance?tenantIdIn="+url.QueryEscape(userId), &result)
	return
}

func getFilteredProcessInstanceHistoryList(userId string, query url.Values) (result HistoricProcessInstances, err error) {
	shard, err := GetUserShard(userId)
	if err != nil {
		return result, err
	}
	query.Del("tenantIdIn")
	err = request.Get(shard+"/engine-rest/history/process-instance?tenantIdIn="+url.QueryEscape(userId)+"&"+query.Encode(), &result)
	return
}

func getProcessInstanceHistoryListFinished(userId string) (result HistoricProcessInstances, err error) {
	shard, err := GetUserShard(userId)
	if err != nil {
		return result, err
	}
	//"/engine-rest/process-instance"
	err = request.Get(shard+"/engine-rest/history/process-instance?tenantIdIn="+url.QueryEscape(userId)+"&finished=true", &result)
	return
}
func getProcessInstanceHistoryListUnfinished(userId string) (result HistoricProcessInstances, err error) {
	shard, err := GetUserShard(userId)
	if err != nil {
		return result, err
	}
	//"/engine-rest/process-instance"
	err = request.Get(shard+"/engine-rest/history/process-instance?tenantIdIn="+url.QueryEscape(userId)+"&unfinished=true", &result)
	return
}
func getProcessInstanceCount(userId string) (result Count, err error) {
	shard, err := GetUserShard(userId)
	if err != nil {
		return result, err
	}
	//"/engine-rest/process-instance/count"
	err = request.Get(shard+"/engine-rest/process-instance/count?tenantIdIn="+url.QueryEscape(userId), &result)
	return
}
func getProcessInstanceList(userId string) (result ProcessInstances, err error) {
	shard, err := GetUserShard(userId)
	if err != nil {
		return result, err
	}
	//"/engine-rest/process-instance"
	err = request.Get(shard+"/engine-rest/process-instance?tenantIdIn="+url.QueryEscape(userId), &result)
	return
}

func getProcessDefinition(id string, userId string) (result ProcessDefinition, err error) {
	shard, err := GetUserShard(userId)
	if err != nil {
		return result, err
	}
	//"/engine-rest/process-definition/" + processDefinitionId
	err = request.Get(shard+"/engine-rest/process-definition/"+url.QueryEscape(id), &result)
	if err != nil {
		return
	}
	err = setVid(&result)
	return
}
func getProcessDefinitionDiagram(id string, userId string) (resp *http.Response, err error) {
	shard, err := GetUserShard(userId)
	if err != nil {
		return resp, err
	}
	// "/engine-rest/process-definition/" + processDefinitionId + "/diagram"
	resp, err = http.Get(shard + "/engine-rest/process-definition/" + url.QueryEscape(id) + "/diagram")
	return
}
func getDeploymentList(userId string, params url.Values) (result Deployments, err error) {
	shard, err := GetUserShard(userId)
	if err != nil {
		return result, err
	}
	// "/engine-rest/deployment?tenantIdIn="+userId
	temp := Deployments{}
	params.Del("tenantIdIn")
	path := shard + "/engine-rest/deployment?tenantIdIn=" + url.QueryEscape(userId) + "&" + params.Encode()
	err = request.Get(path, &temp)
	if err != nil {
		return
	}
	for i := 0; i < len(temp); i++ {
		err = setVid(&temp[i])
		if err != nil {
			log.Println("WARNING: unable to find virtual id for process; ignore process", temp[i].Id, temp[i].Name, err)
		} else {
			result = append(result, temp[i])
		}
	}
	return
}

var UnknownVid = errors.New("unknown vid")
var CamundaDeploymentUnknown = errors.New("deployment unknown in camunda")
var AccessDenied = errors.New("access denied")

func getDefinitionByDeploymentVid(vid string, userId string) (result ProcessDefinitions, err error) {
	id, exists, err := getDeploymentId(vid)
	if err != nil {
		return result, err
	}
	if !exists {
		return result, UnknownVid
	}
	//"/engine-rest/process-definition?deploymentId=
	result, err = getRawDefinitionsByDeployment(id, userId)
	if err != nil {
		return
	}
	for i := 0; i < len(result); i++ {
		err = setVid(&result[i])
		if err != nil {
			return
		}
	}
	return
}

func getRawDefinitionsByDeployment(deploymentId string, userId string) (result ProcessDefinitions, err error) {
	shard, err := GetUserShard(userId)
	if err != nil {
		return result, err
	}
	err = request.Get(shard+"/engine-rest/process-definition?deploymentId="+url.QueryEscape(deploymentId), &result)
	return
}

func getDeployment(vid string, userId string) (result Deployment, err error) {
	shard, err := GetUserShard(userId)
	if err != nil {
		return result, err
	}
	deploymentId, exists, err := getDeploymentId(vid)
	if err != nil {
		return result, err
	}
	if !exists {
		return result, UnknownVid
	}
	//"/engine-rest/deployment/" + id
	err = request.Get(shard+"/engine-rest/deployment/"+url.QueryEscape(deploymentId), &result)
	if err != nil {
		return
	}
	err = setVid(&result)
	return
}

func getDeploymentCountByShard(deploymentId string, shard string) (result Count, err error) {
	err = request.Get(shard+"/engine-rest/deployment/count?id="+url.QueryEscape(deploymentId), &result)
	return
}

func buildPayLoad(name string, xml string, svg string, boundary string, owner string, deploymentSource string) string {
	segments := []string{}
	if deploymentSource == "" {
		deploymentSource = "sepl"
	}

	segments = append(segments, "Content-Disposition: form-data; name=\"data\"; "+"filename=\""+name+".bpmn\"\r\nContent-Type: text/xml\r\n\r\n"+xml+"\r\n")
	segments = append(segments, "Content-Disposition: form-data; name=\"diagram\"; "+"filename=\""+name+".svg\"\r\nContent-Type: image/svg+xml\r\n\r\n"+svg+"\r\n")
	segments = append(segments, "Content-Disposition: form-data; name=\"deployment-name\"\r\n\r\n"+name+"\r\n")
	segments = append(segments, "Content-Disposition: form-data; name=\"deployment-source\"\r\n\r\n"+deploymentSource+"\r\n")
	segments = append(segments, "Content-Disposition: form-data; name=\"tenant-id\"\r\n\r\n"+owner+"\r\n")

	return "--" + boundary + "\r\n" + strings.Join(segments, "--"+boundary+"\r\n") + "--" + boundary + "--\r\n"
}

//returns original deploymentId (not vid)
func DeployProcess(name string, xml string, svg string, owner string, source string) (deploymentId string, err error) {
	responseWrapper, err := deployProcess(name, xml, svg, owner, source)
	if err != nil {
		log.Println("ERROR: unable to decode process engine deployment response", err)
		return deploymentId, err
	}
	ok := false
	deploymentId, ok = responseWrapper["id"].(string)
	if !ok {
		log.Println("ERROR: unable to interpret process engine deployment response", responseWrapper)
		if responseWrapper["type"] == "ProcessEngineException" {
			log.Println("DEBUG: try deploying placeholder process")
			responseWrapper, err = deployProcess(name, createBlankProcess(), createBlankSvg(), owner, source)
			deploymentId, ok = responseWrapper["id"].(string)
			if !ok {
				log.Println("ERROR: unable to deploy placeholder process", responseWrapper)
				err = errors.New("unable to interpret process engine deployment response")
				return
			}
		} else {
			log.Println("ERROR: unable to deploy placeholder process", responseWrapper)
			err = errors.New("unable to interpret process engine deployment response")
			return
		}
	}
	if err == nil && deploymentId == "" {
		err = errors.New("process-engine didnt deploy process: " + xml)
	}
	return
}

func deployProcess(name string, xml string, svg string, owner string, source string) (result map[string]interface{}, err error) {
	shard, err := GetUserShard(owner)
	if err != nil {
		return result, err
	}
	result = map[string]interface{}{}
	boundary := "---------------------------" + time.Now().String()
	b := strings.NewReader(buildPayLoad(name, xml, svg, boundary, owner, source))
	resp, err := http.Post(shard+"/engine-rest/deployment/create", "multipart/form-data; boundary="+boundary, b)
	if err != nil {
		log.Println("ERROR: request to processengine ", err)
		return result, err
	}
	err = json.NewDecoder(resp.Body).Decode(&result)
	return
}

//uses original deploymentId (not vid)
func RemoveProcess(deploymentId string, userId string) (err error) {
	shard, err := GetUserShard(userId)
	if err != nil {
		return err
	}
	return removeProcessForShard(deploymentId, shard)
}

func removeProcessForShard(deploymentId string, shard string) (err error) {
	count, err := getDeploymentCountByShard(deploymentId, shard)
	if err != nil {
		return err
	}
	if count.Count == 0 {
		return nil
	}
	client := &http.Client{}
	url := shard + "/engine-rest/deployment/" + deploymentId + "?cascade=true&skipIoMappings=true"
	request, err := http.NewRequest("DELETE", url, nil)
	_, err = client.Do(request)
	return
}

func RemoveProcessFromAllShards(deploymentId string) (err error) {
	s, err := GetShards()
	if err != nil {
		return err
	}
	shards, err := s.GetShards()
	if err != nil {
		return err
	}
	for _, shard := range shards {
		err = removeProcessForShard(deploymentId, shard)
		if err != nil {
			return err
		}
	}
	return nil
}

func getExtendedDeploymentList(userId string, params url.Values) (result []ExtendedDeployment, err error) {
	deployments, err := getDeploymentList(userId, params)
	if err != nil {
		return result, err
	}
	for _, deployment := range deployments {
		extended, err := getExtendedDeployment(deployment, userId)
		if err != nil {
			result = append(result, ExtendedDeployment{Deployment: deployment, Error: err.Error()})
		} else {
			result = append(result, extended)
		}
	}
	return
}

func getExtendedDeployment(deployment Deployment, userId string) (result ExtendedDeployment, err error) {
	definition, err := getDefinitionByDeploymentVid(deployment.Id, userId)
	if err != nil {
		return result, err
	}
	if len(definition) < 1 {
		return result, errors.New("missing definition for given deployment")
	}
	if len(definition) > 1 {
		return result, errors.New("more than one definition for given deployment")
	}
	svgResp, err := getProcessDefinitionDiagram(definition[0].Id, userId)
	if err != nil {
		return result, err
	}
	svg, err := ioutil.ReadAll(svgResp.Body)
	if err != nil {
		return result, err
	}
	return ExtendedDeployment{Deployment: deployment, Diagram: string(svg), DefinitionId: definition[0].Id}, nil
}

func getProcessInstanceHistoryListWithTotal(userId string, searchtype string, searchvalue string, limit string, offset string, sortby string, sortdirection string, finished bool) (result HistoricProcessInstancesWithTotal, err error) {
	shard, err := GetUserShard(userId)
	if err != nil {
		return result, err
	}
	params := url.Values{
		"tenantIdIn":  []string{userId},
		"maxResults":  []string{limit},
		"firstResult": []string{offset},
		"sortBy":      []string{sortby},
		"sortOrder":   []string{sortdirection},
	}
	if searchtype != "" && searchvalue != "" {
		if searchtype == "processDefinitionId" {
			params["processDefinitionId"] = []string{searchvalue}
		}
		if searchtype == "processDefinitionNameLike" {
			params["processDefinitionNameLike"] = []string{"%" + searchvalue + "%"}
		}

	}
	if finished {
		params["finished"] = []string{"true"}
	} else {
		params["unfinished"] = []string{"true"}
	}

	temp := HistoricProcessInstances{}
	err = request.Get(shard+"/engine-rest/history/process-instance?"+params.Encode(), &temp)
	if err != nil {
		return
	}
	for _, process := range temp {
		result.Data = append(result.Data, process)
	}

	count := Count{}
	err = request.Get(shard+"/engine-rest/history/process-instance/count?"+params.Encode(), &count)
	result.Total = count.Count
	return
}
