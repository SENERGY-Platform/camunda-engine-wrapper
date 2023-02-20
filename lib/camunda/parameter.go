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

package camunda

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/etree"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/camunda/model"
	"io"
	"log"
	"net/http"
	"net/url"
	"runtime/debug"
)

func (this *Camunda) GetProcessParameters(processDefinitionId string, userId string) (result map[string]model.Variable, err error) {
	shard, err := this.shards.EnsureShardForUser(userId)
	if err != nil {
		return result, err
	}
	req, err := http.NewRequest("GET", shard+"/engine-rest/process-definition/"+url.QueryEscape(processDefinitionId)+"/form-variables", nil)
	if err != nil {
		return result, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return result, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		temp, _ := io.ReadAll(resp.Body)
		err = errors.New(resp.Status + " " + string(temp))
		return
	}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return
	}
	return this.filterParameter(shard, processDefinitionId, result)
}

type CamundaXmlWrapper struct {
	Id   string `json:"id"`
	Bpmn string `json:"bpmn20Xml"`
}

func (this *Camunda) getProcessDefinitionXml(shard string, processDefinitionId string) (result CamundaXmlWrapper, err error) {
	req, err := http.NewRequest("GET", shard+"/engine-rest/process-definition/"+url.QueryEscape(processDefinitionId)+"/xml", nil)
	if err != nil {
		return result, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return result, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		temp, _ := io.ReadAll(resp.Body)
		err = errors.New(resp.Status + " " + string(temp))
		return
	}
	err = json.NewDecoder(resp.Body).Decode(&result)
	return
}

func (this *Camunda) filterParameter(shard string, id string, variables map[string]model.Variable) (result map[string]model.Variable, err error) {
	xml, err := this.getProcessDefinitionXml(shard, id)
	if err != nil {
		log.Println("WARNING: unable to filter parameter", err)
		debug.PrintStack()
		return variables, nil //return unfiltered
	}
	parameter, err := this.estimateStartParameter(xml.Bpmn)
	if err != nil {
		log.Println("WARNING: unable to filter parameter", err)
		debug.PrintStack()
		return variables, nil //return unfiltered
	}
	ignoreIndex := map[string]bool{}
	for _, param := range parameter {
		if param.Properties["ignore_on_start"] == "true" {
			ignoreIndex[param.Id] = true
		}
	}
	result = map[string]model.Variable{}
	for key, value := range variables {
		if !ignoreIndex[key] {
			result[key] = value
		}
	}
	return result, nil
}

type ProcessStartParameter struct {
	Id         string            `json:"id"`
	Label      string            `json:"label"`
	Type       string            `json:"type"`
	Default    string            `json:"default"`
	Properties map[string]string `json:"properties"`
}

func (this *Camunda) estimateStartParameter(xml string) (result []ProcessStartParameter, err error) {
	defer func() {
		if r := recover(); r != nil && err == nil {
			log.Printf("%s: %s", r, debug.Stack())
			err = errors.New(fmt.Sprint("Recovered Error: ", r))
		}
	}()
	doc := etree.NewDocument()
	err = doc.ReadFromString(xml)
	if err != nil {
		return
	}
	elements := doc.FindElements("//bpmn:startEvent/bpmn:extensionElements/camunda:formData/camunda:formField")
	for _, element := range elements {
		id := element.SelectAttrValue("id", "")
		if id != "" {
			label := element.SelectAttrValue("label", "")
			paramtype := element.SelectAttrValue("type", "string")
			defaultValue := element.SelectAttrValue("defaultValue", "")
			properties := map[string]string{}
			for _, propterty := range element.FindElements(".//camunda:property") {
				propertyName := propterty.SelectAttrValue("id", "")
				propertyValue := propterty.SelectAttrValue("value", "")
				if propertyName != "" {
					properties[propertyName] = propertyValue
				}
			}
			result = append(result, ProcessStartParameter{
				Id:         id,
				Label:      label,
				Type:       paramtype,
				Default:    defaultValue,
				Properties: properties,
			})
		}
	}
	return result, nil
}
