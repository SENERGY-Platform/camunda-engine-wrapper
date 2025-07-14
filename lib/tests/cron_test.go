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

package tests

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/auth"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/client"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/configuration"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/model"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/tests/helper"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/tests/resources"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/tests/server"
	"io"
	"net/http"
	"slices"
	"sync"
	"testing"
	"time"
)

func TestCron(t *testing.T) {
	t.Log("there is a low probability, that this test fails because the test start lands on second 0 of the current minute while the cron triggers on second 0 of every minute")
	nowTime := time.Now()
	if nowTime.Second() < 5 || nowTime.Second() > 55 {
		time.Sleep(10 * time.Second)
	}
	
	ctx, cancel := context.WithCancel(context.Background())
	wg := sync.WaitGroup{}
	defer wg.Wait()
	defer cancel()

	config, err := configuration.LoadConfig("../../config.json")
	if err != nil {
		t.Error(err)
		return
	}

	config, wrapperUrl, _, err := server.CreateTestEnv(ctx, &wg, config)
	if err != nil {
		t.Error(err)
		return
	}

	token, err := auth.Parse(client.InternalAdminToken)
	if err != nil {
		t.Error(err)
		return
	}

	t.Run("deploy 1", func(t *testing.T) {
		err, _ = client.New(wrapperUrl).Deploy(token.Jwt(), client.DeploymentMessage{
			Deployment: model.Deployment{
				Id:   "test_1",
				Name: "test_1",
				Diagram: model.Diagram{
					XmlDeployed: resources.CronTest,
					Svg:         helper.SvgExample,
				},
				IncidentHandling: &model.IncidentHandling{},
			},
			UserId: token.GetUserId(),
		})
		if err != nil {
			t.Error(err)
			return
		}
	})

	t.Run("deploy 2", func(t *testing.T) {
		err, _ = client.New(wrapperUrl).Deploy(token.Jwt(), client.DeploymentMessage{
			Deployment: model.Deployment{
				Id:   "test_2",
				Name: "test_2",
				Diagram: model.Diagram{
					XmlDeployed: resources.CronTest,
					Svg:         helper.SvgExample,
				},
				IncidentHandling: &model.IncidentHandling{},
			},
			UserId: token.GetUserId(),
		})
		if err != nil {
			t.Error(err)
			return
		}
	})

	t.Run("wait", func(t *testing.T) {
		time.Sleep(time.Minute)
	})

	t.Run("check", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, wrapperUrl+"/v2/history/process-instances?finished=true&with_total=true", nil)
		if err != nil {
			t.Error(err)
			return
		}
		req.Header.Set("Authorization", client.InternalAdminToken)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Error(err)
			return
		}
		defer resp.Body.Close()
		if resp.StatusCode > 299 {
			temp, _ := io.ReadAll(resp.Body) //read error response end ensure that resp.Body is read to EOF
			t.Error(fmt.Errorf("unexpected statuscode %v: %v", resp.StatusCode, string(temp)), resp.StatusCode)
			return
		}
		result := model.HistoricProcessInstancesWithTotal{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		if err != nil {
			t.Error(err)
			return
		}
		t.Log(resp.StatusCode, len(result.Data), result.Total)
		if len(result.Data) != 2 {
			t.Error(len(result.Data))
			t.Errorf("%#v", result)
			return
		}
		if slices.IndexFunc(result.Data, func(instance model.HistoricProcessInstance) bool {
			return instance.ProcessDefinitionKey == "deplid_test_1"
		}) == -1 {
			t.Error("deplid_test_1 not found")
			t.Errorf("%#v", result)
			return
		}
		if slices.IndexFunc(result.Data, func(instance model.HistoricProcessInstance) bool {
			return instance.ProcessDefinitionKey == "deplid_test_2"
		}) == -1 {
			t.Error("deplid_test_2 not found")
			t.Errorf("%#v", result)
			return
		}
	})

}
