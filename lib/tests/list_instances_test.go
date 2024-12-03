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

package tests

import (
	"context"
	"encoding/json"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/camunda/model"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/configuration"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/tests/helper"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/tests/server"
	"io"
	"net/http"
	"sync"
	"testing"
)

func TestListInstances(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	wg := sync.WaitGroup{}
	defer wg.Wait()
	defer cancel()

	config, err := configuration.LoadConfig("../../config.json")
	if err != nil {
		t.Error(err)
		return
	}

	config, wrapperUrl, shard, e, err := server.CreateTestEnv(ctx, &wg, config)
	if err != nil {
		t.Error(err)
		return
	}

	deploymentId := "withInput"
	t.Run("deploy process with input", testDeployProcessWithInput(e, deploymentId, processWithInput))

	t.Run("start process with input inputTemperature 30", testStartProcessWithInput(wrapperUrl, deploymentId, map[string]interface{}{"inputTemperature": 30}))
	t.Run("start process with input inputTemperature 21", testStartProcessWithInput(wrapperUrl, deploymentId, map[string]interface{}{"inputTemperature": 21}))
	t.Run("start process with partially unused inputs", testStartProcessWithInput(wrapperUrl, deploymentId, map[string]interface{}{"inputTemperature": 10, "unused": "foo"}))

	t.Run("fetch and complete one task to finish one process", testFetchAndComplete(shard))

	t.Run("deploy process without input", testDeployProcessWithInput(e, "withoutInput", processWithoutInput))
	t.Run("start process without inputs", testStartProcessWithInput(wrapperUrl, "withoutInput", nil))

	t.Run("list first deployment instances", testListDeploymentInstances(wrapperUrl, deploymentId, 2))
	t.Run("list second deployment instances", testListDeploymentInstances(wrapperUrl, "withoutInput", 1))

	t.Run("delete unknown instance", func(t *testing.T) {
		req, err := http.NewRequest("DELETE", wrapperUrl+"/v2/process-instances/unknown", nil)
		if err != nil {
			t.Error(err)
			return
		}
		req.Header.Set("Authorization", helper.Jwt)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Error(err)
			return
		}
		defer resp.Body.Close()
		temp, _ := io.ReadAll(resp.Body)
		if resp.StatusCode != 200 && resp.StatusCode != 404 {
			t.Error(resp.StatusCode, string(temp))
			return
		}
	})
}

func testListDeploymentInstances(wrapper string, id string, expectedCount int) func(t *testing.T) {
	return func(t *testing.T) {
		req, err := http.NewRequest("GET", wrapper+"/v2/deployments/"+id+"/instances", nil)
		if err != nil {
			t.Error(err)
			return
		}
		req.Header.Set("Authorization", string(helper.Jwt))
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Error(err)
			return
		}
		defer resp.Body.Close()
		temp, _ := io.ReadAll(resp.Body)
		if resp.StatusCode != 200 {
			t.Error(resp.StatusCode, string(temp))
			return
		}
		result := model.ProcessInstances{}
		err = json.Unmarshal(temp, &result)
		if resp.StatusCode != 200 {
			t.Error(resp.StatusCode, string(temp))
			return
		}
		if len(result) != expectedCount {
			t.Error(len(result), "\n", string(temp))
			return
		}
	}
}

func testFetchAndComplete(url string) func(t *testing.T) {
	return func(t *testing.T) {
		tasks, err := fetchTestTask(url)
		if err != nil {
			t.Error(err)
			return
		}
		err = completeTask(url, tasks[0].Id)
		if err != nil {
			t.Error(err)
			return
		}
	}
}
