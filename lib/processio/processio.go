/*
 * Copyright 2022 InfAI (CC SES)
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

package processio

import (
	"errors"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/auth"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/configuration"
	"io"
	"net/http"
	"net/url"
	"runtime/debug"
)

type ProcessIo struct {
	config      configuration.Config
	adminAccess *auth.OpenidToken
}

func NewOrNil(config configuration.Config) *ProcessIo {
	if config.ProcessIoUrl == "" {
		return nil
	}
	return New(config)
}

func New(config configuration.Config) *ProcessIo {
	return &ProcessIo{
		config:      config,
		adminAccess: &auth.OpenidToken{},
	}
}

func (this *ProcessIo) DeleteProcessDefinition(definitionId string) error {
	token, err := this.adminAccess.EnsureAccess(this.config)
	if err != nil {
		debug.PrintStack()
		return err
	}
	req, err := http.NewRequest("DELETE", this.config.ProcessIoUrl+"/process-definitions/"+url.PathEscape(definitionId), nil)
	if err != nil {
		debug.PrintStack()
		return err
	}
	req.Header.Set("Authorization", token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		debug.PrintStack()
		return err
	}
	defer resp.Body.Close()
	respMsg, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		debug.PrintStack()
		return errors.New(string(respMsg))
	}
	return nil
}

func (this *ProcessIo) DeleteProcessInstance(instanceId string) error {
	token, err := this.adminAccess.EnsureAccess(this.config)
	if err != nil {
		debug.PrintStack()
		return err
	}
	req, err := http.NewRequest("DELETE", this.config.ProcessIoUrl+"/process-instances/"+url.PathEscape(instanceId), nil)
	if err != nil {
		debug.PrintStack()
		return err
	}
	req.Header.Set("Authorization", token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		debug.PrintStack()
		return err
	}
	defer resp.Body.Close()
	respMsg, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		debug.PrintStack()
		return errors.New(string(respMsg))
	}
	return nil
}
