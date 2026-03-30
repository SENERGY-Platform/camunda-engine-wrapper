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

package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/model"
)

type Client struct {
	serverUrl string
}

func New(serverUrl string) (client *Client) {
	return &Client{serverUrl: serverUrl}
}

type DeploymentMessage = model.DeploymentMessage
type Deployment = model.Deployment
type Diagram = model.Diagram
type IncidentHandling = model.IncidentHandling
type ProcessInstance = model.ProcessInstance
type HistoricProcessInstances = model.HistoricProcessInstances
type HistoricProcessInstancesWithTotal = model.HistoricProcessInstancesWithTotal
type ExtendedDeployment = model.ExtendedDeployment

type StartOptions struct {
	BusinessKey string
	Inputs      map[string]interface{}
}

type InstanceListOptions struct {
	BusinessKey string
	OtherArgs   map[string]string
}

type DeploymentListOptions struct {
	OtherArgs map[string]string
}

func (this *Client) Deploy(token string, depl DeploymentMessage) (err error, code int) {
	body, err := json.Marshal(depl)
	if err != nil {
		return err, 0
	}
	req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("%v/process-deployments", this.serverUrl), bytes.NewBuffer(body))
	if err != nil {
		return err, 0
	}
	return doVoid(token, req)
}

func (this *Client) ListDeployments(token string, options DeploymentListOptions) (result []ExtendedDeployment, err error, code int) {
	query := url.Values{}
	for key, val := range options.OtherArgs {
		query.Add(key, val)
	}
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%v/v2/deployments?%v", this.serverUrl, query.Encode()), nil)
	if err != nil {
		return result, err, 0
	}
	return do[[]ExtendedDeployment](token, req)
}

func (this *Client) DeleteDeployment(token string, userId string, deplId string) (err error, code int) {
	req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("%v/process-deployments/%v/%v", this.serverUrl, userId, deplId), nil)
	if err != nil {
		return err, 0
	}
	return doVoid(token, req)
}

func (this *Client) GetProcessInstances(token string) (result HistoricProcessInstances, err error, code int) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%v/v2/process-instances", this.serverUrl), nil)
	if err != nil {
		return result, err, 0
	}
	return do[HistoricProcessInstances](token, req)
}

func (this *Client) GetHistoricProcessInstances(token string, options InstanceListOptions) (result HistoricProcessInstances, err error, code int) {
	query := url.Values{}
	if options.BusinessKey != "" {
		query.Add("processInstanceBusinessKey", options.BusinessKey)
	}
	for key, val := range options.OtherArgs {
		query.Add(key, val)
	}
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%v/v2/history/process-instances?%v", this.serverUrl, query.Encode()), nil)
	if err != nil {
		return result, err, 0
	}
	return do[HistoricProcessInstances](token, req)
}

func (this *Client) GetHistoricProcessInstancesWithTotal(token string, options InstanceListOptions) (result HistoricProcessInstancesWithTotal, err error, code int) {
	query := url.Values{}
	query.Add("with_total", "true")
	if options.BusinessKey != "" {
		query.Add("processInstanceBusinessKey", options.BusinessKey)
	}
	for key, val := range options.OtherArgs {
		query.Add(key, val)
	}
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%v/v2/history/process-instances?%v", this.serverUrl, query.Encode()), nil)
	if err != nil {
		return result, err, 0
	}
	return do[HistoricProcessInstancesWithTotal](token, req)
}

func (this *Client) StartDeployment(token string, deplId string, options StartOptions) (result ProcessInstance, err error, code int) {
	query := url.Values{}
	if options.BusinessKey != "" {
		query.Add("business_key", options.BusinessKey)
	}
	for key, val := range options.Inputs {
		str, _ := json.Marshal(val)
		query.Add(key, string(str))
	}
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%v/v2/deployments/%v/start?%v", this.serverUrl, url.PathEscape(deplId), query.Encode()), nil)
	if err != nil {
		return result, err, 0
	}
	return do[ProcessInstance](token, req)
}

func (this *Client) DeleteProcessInstancesByBusinessKey(token string, businessKey string) (err error, code int) {
	req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("%v/v2/process-instances-by-business-key/%v", this.serverUrl, url.PathEscape(businessKey)), nil)
	if err != nil {
		return err, 0
	}
	return doVoid(token, req)
}

func do[T any](token string, req *http.Request) (result T, err error, code int) {
	req.Header.Set("Authorization", token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	defer resp.Body.Close()
	if resp.StatusCode > 299 {
		temp, _ := io.ReadAll(resp.Body) //read error response end ensure that resp.Body is read to EOF
		return result, fmt.Errorf("unexpected statuscode %v: %v", resp.StatusCode, string(temp)), resp.StatusCode
	}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		_, _ = io.ReadAll(resp.Body) //ensure resp.Body is read to EOF
		return result, err, http.StatusInternalServerError
	}
	return result, nil, resp.StatusCode
}

func doVoid(token string, req *http.Request) (err error, code int) {
	req.Header.Set("Authorization", token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	defer resp.Body.Close()
	if resp.StatusCode > 299 {
		temp, _ := io.ReadAll(resp.Body) //read error response end ensure that resp.Body is read to EOF
		return fmt.Errorf("unexpected statuscode %v: %v", resp.StatusCode, string(temp)), resp.StatusCode
	}
	return nil, resp.StatusCode
}
