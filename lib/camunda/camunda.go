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

package camunda

import (
	"bytes"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/configuration"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/model"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/notification"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/processio"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/shards"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/vid"
	"io"
	"io/ioutil"
	"net/url"
	"strconv"
	"strings"
	"time"

	"errors"
	"net/http"

	"encoding/json"

	"log"
)

type Camunda struct {
	shards    *shards.Shards
	vid       *vid.Vid
	config    configuration.Config
	processIo *processio.ProcessIo
}

func New(config configuration.Config, vid *vid.Vid, shards *shards.Shards, processIo *processio.ProcessIo) *Camunda {
	return &Camunda{config: config, vid: vid, shards: shards, processIo: processIo}
}

func (this *Camunda) StartProcess(processDefinitionId string, userId string, parameter map[string]interface{}) (err error) {
	shard, err := this.shards.EnsureShardForUser(userId)
	if err != nil {
		return err
	}

	message := createStartMessage(parameter)

	b := new(bytes.Buffer)
	err = json.NewEncoder(b).Encode(message)
	if err != nil {
		return
	}
	req, err := http.NewRequest("POST", shard+"/engine-rest/process-definition/"+url.QueryEscape(processDefinitionId)+"/submit-form", b)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	temp, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		err = errors.New(resp.Status + " " + string(temp))
		return
	}
	return nil
}

func createStartMessage(parameter map[string]interface{}) map[string]interface{} {
	if len(parameter) == 0 {
		return map[string]interface{}{}
	}
	variables := map[string]interface{}{}
	for key, val := range parameter {
		variables[key] = map[string]interface{}{
			"value": val,
		}
	}
	return map[string]interface{}{"variables": variables}
}

func (this *Camunda) StartProcessGetId(processDefinitionId string, userId string, parameter map[string]interface{}) (result model.ProcessInstance, err error) {
	shard, err := this.shards.EnsureShardForUser(userId)
	if err != nil {
		return result, err
	}

	message := createStartMessage(parameter)

	b := new(bytes.Buffer)
	err = json.NewEncoder(b).Encode(message)
	if err != nil {
		return
	}
	req, err := http.NewRequest("POST", shard+"/engine-rest/process-definition/"+url.QueryEscape(processDefinitionId)+"/submit-form", b)
	if err != nil {
		return result, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return result, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		temp, _ := ioutil.ReadAll(resp.Body)
		err = errors.New(resp.Status + " " + string(temp))
		return
	}
	err = json.NewDecoder(resp.Body).Decode(&result)
	return
}

func (this *Camunda) CheckProcessDefinitionAccess(id string, userId string) (err error) {
	shard, err := this.shards.EnsureShardForUser(userId)
	if err != nil {
		return err
	}
	definition := model.ProcessDefinition{}
	err = Get(shard+"/engine-rest/process-definition/"+url.QueryEscape(id), &definition)
	if err == nil && definition.TenantId != userId {
		err = errors.New("access denied")
	}
	return
}

func (this *Camunda) CheckDeploymentAccess(vid string, userId string) (err error) {
	shard, err := this.shards.EnsureShardForUser(userId)
	if err != nil {
		return err
	}
	id, exists, err := this.vid.GetDeploymentId(vid)
	if err != nil {
		return err
	}
	if !exists {
		return UnknownVid
	}
	wrapper := model.CamundaDeployment{}
	err = Get(shard+"/engine-rest/deployment/"+url.QueryEscape(id), &wrapper)
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

var ErrAccessDenied = errors.New("access denied")

func (this *Camunda) CheckProcessInstanceAccess(id string, userId string) (err error, code int) {
	shard, err := this.shards.EnsureShardForUser(userId)
	if err != nil {
		return err, 500
	}
	resp, err := http.Get(shard + "/engine-rest/process-instance/" + url.QueryEscape(id))
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		err = errors.New(resp.Status + " " + string(b))
		return err, resp.StatusCode
	}
	wrapper := model.ProcessInstance{}
	err = json.NewDecoder(resp.Body).Decode(&wrapper)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	if wrapper.TenantId != userId {
		return ErrAccessDenied, http.StatusForbidden
	}
	return nil, http.StatusOK
}

func (this *Camunda) CheckHistoryAccess(id string, userId string) (definitionId string, err error) {
	shard, err := this.shards.EnsureShardForUser(userId)
	if err != nil {
		return definitionId, err
	}
	wrapper := model.HistoricProcessInstance{}
	err = Get(shard+"/engine-rest/history/process-instance/"+url.QueryEscape(id), &wrapper)
	if err == nil && wrapper.TenantId != userId {
		err = errors.New("access denied")
	}
	return wrapper.ProcessDefinitionId, err
}

func (this *Camunda) RemoveProcessInstance(id string, userId string) (err error) {
	if this.processIo != nil {
		err = this.processIo.DeleteProcessInstance(id)
		if err != nil {
			return err
		}
	}

	////DELETE "/engine-rest/process-instance/" + processInstanceId
	shard, err := this.shards.EnsureShardForUser(userId)
	if err != nil {
		return err
	}
	request, err := http.NewRequest("DELETE", shard+"/engine-rest/process-instance/"+url.QueryEscape(id)+"?skipIoMappings=true", nil)
	if err != nil {
		return
	}
	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	msg, _ := io.ReadAll(resp.Body)
	if !(resp.StatusCode == 200 || resp.StatusCode == 204) {
		u, _ := url.Parse(shard)
		u.User = &url.Userinfo{}
		err = errors.New("error on delete in engine for " + u.String() + "/engine-rest/process-instance/" + url.QueryEscape(id) + ": " + resp.Status + " " + string(msg))
	}
	return
}

func (this *Camunda) RemoveProcessInstanceHistory(id string, userId string) (err error) {
	if this.processIo != nil {
		err = this.processIo.DeleteProcessInstance(id)
		if err != nil {
			return err
		}
	}

	shard, err := this.shards.EnsureShardForUser(userId)
	if err != nil {
		return err
	}
	//DELETE "/engine-rest/history/process-instance/" + processInstanceId
	u, _ := url.Parse(shard)
	u.User = &url.Userinfo{}
	request, err := http.NewRequest("DELETE", shard+"/engine-rest/history/process-instance/"+url.QueryEscape(id), nil)
	if err != nil {
		return
	}
	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if err == nil && !(resp.StatusCode == 200 || resp.StatusCode == 204) {
		msg, _ := ioutil.ReadAll(resp.Body)
		err = errors.New("error on delete in engine for " + u.String() + "/engine-rest/history/process-instance/" + url.QueryEscape(id) + ": " + resp.Status + " " + string(msg))
	}
	return
}

func (this *Camunda) GetProcessInstanceHistoryByProcessDefinition(id string, userId string) (result model.HistoricProcessInstances, err error) {
	shard, err := this.shards.EnsureShardForUser(userId)
	if err != nil {
		return result, err
	}
	//"/engine-rest/history/process-instance?processDefinitionId="
	err = Get(shard+"/engine-rest/history/process-instance?processDefinitionId="+url.QueryEscape(id), &result)
	return
}
func (this *Camunda) GetProcessInstanceHistoryByProcessDefinitionFinished(id string, userId string) (result model.HistoricProcessInstances, err error) {
	shard, err := this.shards.EnsureShardForUser(userId)
	if err != nil {
		return result, err
	}
	//"/engine-rest/history/process-instance?processDefinitionId="
	err = Get(shard+"/engine-rest/history/process-instance?processDefinitionId="+url.QueryEscape(id)+"&finished=true", &result)
	return
}
func (this *Camunda) GetProcessInstanceHistoryByProcessDefinitionUnfinished(id string, userId string) (result model.HistoricProcessInstances, err error) {
	shard, err := this.shards.EnsureShardForUser(userId)
	if err != nil {
		return result, err
	}
	//"/engine-rest/history/process-instance?processDefinitionId="
	err = Get(shard+"/engine-rest/history/process-instance?processDefinitionId="+url.QueryEscape(id)+"&unfinished=true", &result)
	return
}

func (this *Camunda) GetProcessInstanceHistoryList(userId string) (result model.HistoricProcessInstances, err error) {
	shard, err := this.shards.EnsureShardForUser(userId)
	if err != nil {
		return result, err
	}
	//"/engine-rest/process-instance"
	err = Get(shard+"/engine-rest/history/process-instance?tenantIdIn="+url.QueryEscape(userId), &result)
	return
}

func (this *Camunda) GetFilteredProcessInstanceHistoryList(userId string, query url.Values) (result model.HistoricProcessInstances, err error) {
	shard, err := this.shards.EnsureShardForUser(userId)
	if err != nil {
		return result, err
	}
	query.Del("tenantIdIn")
	err = Get(shard+"/engine-rest/history/process-instance?tenantIdIn="+url.QueryEscape(userId)+"&"+query.Encode(), &result)
	return
}

func (this *Camunda) GetFilteredProcessInstanceHistoryListWithTotal(userId string, query url.Values) (result model.HistoricProcessInstancesWithTotal, err error) {
	shard, err := this.shards.EnsureShardForUser(userId)
	if err != nil {
		return result, err
	}
	query.Del("tenantIdIn")
	err = Get(shard+"/engine-rest/history/process-instance?tenantIdIn="+url.QueryEscape(userId)+"&"+query.Encode(), &result.Data)
	if err != nil {
		return result, err
	}
	count := model.Count{}
	err = Get(shard+"/engine-rest/history/process-instance/count?tenantIdIn="+url.QueryEscape(userId)+"&"+query.Encode(), &count)
	result.Total = count.Count
	return
}

func (this *Camunda) GetProcessInstanceHistoryListFinished(userId string) (result model.HistoricProcessInstances, err error) {
	shard, err := this.shards.EnsureShardForUser(userId)
	if err != nil {
		return result, err
	}
	//"/engine-rest/process-instance"
	err = Get(shard+"/engine-rest/history/process-instance?tenantIdIn="+url.QueryEscape(userId)+"&finished=true", &result)
	return
}
func (this *Camunda) GetProcessInstanceHistoryListUnfinished(userId string) (result model.HistoricProcessInstances, err error) {
	shard, err := this.shards.EnsureShardForUser(userId)
	if err != nil {
		return result, err
	}
	//"/engine-rest/process-instance"
	err = Get(shard+"/engine-rest/history/process-instance?tenantIdIn="+url.QueryEscape(userId)+"&unfinished=true", &result)
	return
}
func (this *Camunda) GetProcessInstanceCount(userId string) (result model.Count, err error) {
	shard, err := this.shards.EnsureShardForUser(userId)
	if err != nil {
		return result, err
	}
	//"/engine-rest/process-instance/count"
	err = Get(shard+"/engine-rest/process-instance/count?tenantIdIn="+url.QueryEscape(userId), &result)
	return
}
func (this *Camunda) GetProcessInstanceList(userId string) (result model.ProcessInstances, err error) {
	shard, err := this.shards.EnsureShardForUser(userId)
	if err != nil {
		return result, err
	}
	//"/engine-rest/process-instance"
	err = Get(shard+"/engine-rest/process-instance?tenantIdIn="+url.QueryEscape(userId), &result)
	return
}

func (this *Camunda) GetProcessDefinition(id string, userId string) (result model.ProcessDefinition, err error) {
	shard, err := this.shards.EnsureShardForUser(userId)
	if err != nil {
		return result, err
	}
	//"/engine-rest/process-definition/" + processDefinitionId
	err = Get(shard+"/engine-rest/process-definition/"+url.QueryEscape(id), &result)
	if err != nil {
		return
	}
	err = this.vid.SetVid(&result)
	return
}
func (this *Camunda) GetProcessDefinitionDiagram(id string, userId string) (resp *http.Response, err error) {
	shard, err := this.shards.EnsureShardForUser(userId)
	if err != nil {
		return resp, err
	}
	// "/engine-rest/process-definition/" + processDefinitionId + "/diagram"
	resp, err = http.Get(shard + "/engine-rest/process-definition/" + url.QueryEscape(id) + "/diagram")
	return
}
func (this *Camunda) GetDeploymentList(userId string, params url.Values) (result model.CamundaDeployments, err error) {
	shard, err := this.shards.EnsureShardForUser(userId)
	if err != nil {
		return result, err
	}
	// "/engine-rest/deployment?tenantIdIn="+userId
	temp := model.CamundaDeployments{}
	params.Del("tenantIdIn")
	path := shard + "/engine-rest/deployment?tenantIdIn=" + url.QueryEscape(userId) + "&" + params.Encode()
	err = Get(path, &temp)
	if err != nil {
		return
	}
	for i := 0; i < len(temp); i++ {
		err = this.vid.SetVid(&temp[i])
		if err != nil {
			log.Println("WARNING: unable to find virtual id for process; ignore process", temp[i].Id, temp[i].Name, err)
			err = nil
		} else {
			result = append(result, temp[i])
		}
	}
	return
}

var UnknownVid = errors.New("unknown vid")
var CamundaDeploymentUnknown = errors.New("deployment unknown in camunda")
var AccessDenied = errors.New("access denied")

func (this *Camunda) GetDefinitionByDeploymentVid(vid string, userId string) (result model.ProcessDefinitions, err error) {
	id, exists, err := this.vid.GetDeploymentId(vid)
	if err != nil {
		return result, err
	}
	if !exists {
		return result, UnknownVid
	}
	//"/engine-rest/process-definition?deploymentId=
	result, err = this.GetRawDefinitionsByDeployment(id, userId)
	if err != nil {
		return
	}
	for i := 0; i < len(result); i++ {
		err = this.vid.SetVid(&result[i])
		if err != nil {
			return
		}
	}
	return
}

func (this *Camunda) GetInstancesByDeploymentVid(vid string, userId string) (result model.ProcessInstances, err error) {
	id, exists, err := this.vid.GetDeploymentId(vid)
	if err != nil {
		return result, err
	}
	if !exists {
		return result, UnknownVid
	}

	shard, err := this.shards.EnsureShardForUser(userId)
	if err != nil {
		return result, err
	}
	err = Get(shard+"/engine-rest/process-instance?tenantIdIn="+url.QueryEscape(userId)+"&deploymentId="+url.QueryEscape(id), &result)
	return
}

func (this *Camunda) GetRawDefinitionsByDeployment(deploymentId string, userId string) (result model.ProcessDefinitions, err error) {
	shard, err := this.shards.EnsureShardForUser(userId)
	if err != nil {
		return result, err
	}
	err = Get(shard+"/engine-rest/process-definition?deploymentId="+url.QueryEscape(deploymentId), &result)
	return
}

func (this *Camunda) GetDeployment(vid string, userId string) (result model.CamundaDeployment, err error) {
	shard, err := this.shards.EnsureShardForUser(userId)
	if err != nil {
		return result, err
	}
	deploymentId, exists, err := this.vid.GetDeploymentId(vid)
	if err != nil {
		return result, err
	}
	if !exists {
		return result, UnknownVid
	}
	//"/engine-rest/deployment/" + id
	err = Get(shard+"/engine-rest/deployment/"+url.QueryEscape(deploymentId), &result)
	if err != nil {
		return
	}
	err = this.vid.SetVid(&result)
	return
}

func (this *Camunda) GetDeploymentCountByShard(deploymentId string, shard string) (result model.Count, err error) {
	err = Get(shard+"/engine-rest/deployment/count?id="+url.QueryEscape(deploymentId), &result)
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

// returns original deploymentId (not vid)
func (this *Camunda) DeployProcess(name string, xml string, svg string, owner string, source string) (deploymentId string, err error) {
	responseWrapper, err := this.deployProcess(name, xml, svg, owner, source)
	if err != nil {
		log.Println("ERROR: unable to decode process engine deployment response", err)
		return deploymentId, err
	}
	ok := false
	deploymentId, ok = responseWrapper["id"].(string)
	if !ok {
		log.Println("ERROR: unable to interpret process engine deployment response", responseWrapper)
		if responseWrapper["type"] == "ProcessEngineException" {
			msg, ok := responseWrapper["message"].(string)
			if !ok {
				msg = ""
			}
			_ = notification.Send(this.config.NotificationUrl, notification.Message{
				UserId:  owner,
				Title:   "Deployment Error: ProcessEngineException",
				Message: msg,
			})
			log.Println("DEBUG: try deploying placeholder process")
			responseWrapper, err = this.deployProcess(name, CreateBlankProcess(), CreateBlankSvg(), owner, source)
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

func (this *Camunda) deployProcess(name string, xml string, svg string, owner string, source string) (result map[string]interface{}, err error) {
	shard, err := this.shards.EnsureShardForUser(owner)
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

// uses original deploymentId (not vid)
func (this *Camunda) RemoveProcess(deploymentId string, userId string) (err error) {
	shard, err := this.shards.EnsureShardForUser(userId)
	if err != nil {
		return err
	}
	return this.RemoveProcessForShard(deploymentId, shard)
}

func (this *Camunda) RemoveProcessForShard(deploymentId string, shard string) (err error) {
	count, err := this.GetDeploymentCountByShard(deploymentId, shard)
	if err != nil {
		return err
	}
	if count.Count == 0 {
		return nil
	}
	url := shard + "/engine-rest/deployment/" + deploymentId + "?cascade=true&skipIoMappings=true"
	request, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	payload, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 && resp.StatusCode != 404 {
		log.Println("ERROR:", resp.Status, string(payload))
		err = errors.New(resp.Status)
	}
	return err
}

func (this *Camunda) RemoveProcessFromAllShards(deploymentId string) (err error) {
	shards, err := this.shards.GetShards()
	if err != nil {
		return err
	}
	for _, shard := range shards {
		err = this.RemoveProcessForShard(deploymentId, shard)
		if err != nil {
			return err
		}
	}
	return nil
}

func (this *Camunda) GetExtendedDeploymentList(userId string, params url.Values) (result []model.ExtendedDeployment, err error) {
	deployments, err := this.GetDeploymentList(userId, params)
	if err != nil {
		return result, err
	}
	for _, deployment := range deployments {
		extended, err := this.GetExtendedDeployment(deployment, userId)
		if err != nil {
			result = append(result, model.ExtendedDeployment{CamundaDeployment: deployment, Error: err.Error()})
			err = nil
		} else {
			result = append(result, extended)
		}
	}
	return
}

func (this *Camunda) GetExtendedDeployment(deployment model.CamundaDeployment, userId string) (result model.ExtendedDeployment, err error) {
	definition, err := this.GetDefinitionByDeploymentVid(deployment.Id, userId)
	if err != nil {
		return result, err
	}
	if len(definition) < 1 {
		return result, errors.New("missing definition for given deployment")
	}
	if len(definition) > 1 {
		return result, errors.New("more than one definition for given deployment")
	}
	svgResp, err := this.GetProcessDefinitionDiagram(definition[0].Id, userId)
	if err != nil {
		return result, err
	}
	svg, err := io.ReadAll(svgResp.Body)
	if err != nil {
		return result, err
	}
	return model.ExtendedDeployment{CamundaDeployment: deployment, Diagram: string(svg), DefinitionId: definition[0].Id}, nil
}

func (this *Camunda) GetProcessInstanceHistoryListWithTotal(userId string, searchtype string, searchvalue string, limit string, offset string, sortby string, sortdirection string, finished bool) (result model.HistoricProcessInstancesWithTotal, err error) {
	shard, err := this.shards.EnsureShardForUser(userId)
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

	temp := model.HistoricProcessInstances{}
	err = Get(shard+"/engine-rest/history/process-instance?"+params.Encode(), &temp)
	if err != nil {
		return
	}
	for _, process := range temp {
		result.Data = append(result.Data, process)
	}

	count := model.Count{}
	err = Get(shard+"/engine-rest/history/process-instance/count?"+params.Encode(), &count)
	result.Total = count.Count
	return
}

func (this *Camunda) SendEventTrigger(userId string, msg map[string]interface{}) (response []byte, err error) {
	shard, err := this.shards.EnsureShardForUser(userId)
	if err != nil {
		return response, err
	}
	//ensure userId in message
	msg["tenantId"] = userId
	requestWIthUserId, err := json.Marshal(msg)
	if err != nil {
		return response, err
	}

	log.Printf("trigger event: %#v\n", msg)
	resp, err := http.Post(shard+"/engine-rest/message", "application/json", bytes.NewBuffer(requestWIthUserId))
	if err != nil {
		return response, err
	}
	defer resp.Body.Close()
	response, err = io.ReadAll(resp.Body)
	if err != nil {
		return response, err
	}
	if resp.StatusCode >= 300 {
		err = errors.New(string(response))
	}
	return response, err
}

func (this *Camunda) SetProcessInstanceVariable(instanceId string, userId string, variableName string, variableValue interface{}) error {
	shard, err := this.shards.EnsureShardForUser(userId)
	if err != nil {
		return err
	}

	msg := map[string]interface{}{
		"modifications": map[string]interface{}{
			variableName: map[string]interface{}{
				"value": variableValue,
			},
		},
	}
	b := new(bytes.Buffer)
	err = json.NewEncoder(b).Encode(msg)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", shard+"/engine-rest/process-instance/"+url.PathEscape(instanceId)+"/variables", b)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	temp, _ := io.ReadAll(resp.Body)

	if resp.StatusCode >= 300 {
		err = errors.New(resp.Status + " " + string(temp))
		return err
	}

	return nil
}

func CreateBlankSvg() string {
	return `<svg xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" version="1.2" id="Layer_1" x="0px" y="0px" viewBox="0 0 20 16" xml:space="preserve">
<path fill="#D61F33" d="M10,0L0,16h20L10,0z M11,13.908H9v-2h2V13.908z M9,10.908v-6h2v6H9z"/>
</svg>`
}

func CreateBlankProcess() string {
	templ := `<bpmn:definitions xmlns:xsi='http://www.w3.org/2001/XMLSchema-instance' xmlns:bpmn='http://www.omg.org/spec/BPMN/20100524/MODEL' xmlns:bpmndi='http://www.omg.org/spec/BPMN/20100524/DI' xmlns:dc='http://www.omg.org/spec/DD/20100524/DC' id='Definitions_1' targetNamespace='http://bpmn.io/schema/bpmn'><bpmn:process id='PROCESSID' isExecutable='true'><bpmn:startEvent id='StartEvent_1'/></bpmn:process><bpmndi:BPMNDiagram id='BPMNDiagram_1'><bpmndi:BPMNPlane id='BPMNPlane_1' bpmnElement='PROCESSID'><bpmndi:BPMNShape id='_BPMNShape_StartEvent_2' bpmnElement='StartEvent_1'><dc:Bounds x='173' y='102' width='36' height='36'/></bpmndi:BPMNShape></bpmndi:BPMNPlane></bpmndi:BPMNDiagram></bpmn:definitions>`
	return strings.Replace(templ, "PROCESSID", "id_"+strconv.FormatInt(time.Now().Unix(), 10), 1)
}
