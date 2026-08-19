/*
 * Copyright 2026 InfAI (CC SES)
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
	"fmt"
	"io"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/client"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/configuration"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/tests/docker"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/tests/helper"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/tests/server"
)

func TestMigration(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	wg := sync.WaitGroup{}
	defer wg.Wait()
	defer cancel()

	config, err := configuration.LoadConfig("../../config.json")
	if err != nil {
		t.Error(err)
		return
	}

	pgStr, err := docker.Postgres(ctx, &wg, "vid_relations")
	if err != nil {
		t.Error(err)
		return
	}

	camundaCtx, camundaCancel := context.WithCancel(ctx)

	_, camundaPgIp, _, err := docker.PostgresWithNetwork(ctx, &wg, "camunda")
	if err != nil {
		t.Error(err)
		return
	}

	config.WrapperDb = pgStr
	config.ShardingDb = pgStr
	config.CamundaDb = fmt.Sprintf("postgres://usr:pw@%s:5432/camunda?sslmode=disable", camundaPgIp)

	config, wrapperUrl, shard, err := server.CreateTestEnvWithCamundaDatabaseAndTag(camundaCtx, &wg, config, camundaPgIp, "v1.0.4")
	if err != nil {
		t.Error(err)
		return
	}

	wrapperClient := client.New(wrapperUrl)

	deploymentId := "withInput"
	t.Run("deploy process with input", testDeployProcessWithInput(wrapperClient, deploymentId, processWithInput))

	t.Run("start process with input inputTemperature 30", testStartProcessWithInput(wrapperUrl, deploymentId, map[string]interface{}{"inputTemperature": 30, "business_key": "bk1"}))
	t.Run("start process with input inputTemperature 21", testStartProcessWithInput(wrapperUrl, deploymentId, map[string]interface{}{"inputTemperature": 21}))
	t.Run("start process with partially unused inputs", testStartProcessWithInput(wrapperUrl, deploymentId, map[string]interface{}{"inputTemperature": 10, "unused": "foo", "business_key": "bk2"}))

	t.Run("update camunda", func(t *testing.T) {
		camundaCancel()

		config, wrapperUrl, shard, err = server.CreateTestEnvWithCamundaDatabaseAndTag(ctx, &wg, config, camundaPgIp, "v1.0.5")
		if err != nil {
			t.Error(err)
			return
		}

		wrapperClient = client.New(wrapperUrl)
		time.Sleep(10 * time.Second)
	})

	t.Run("fetch and complete one task to finish one process", testFetchAndComplete(shard))

	t.Run("deploy process without input", testDeployProcessWithInput(wrapperClient, "withoutInput", processWithoutInput))
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
