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

package api

import (
	"encoding/json"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/camunda"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/camunda/model"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/configuration"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/events"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/shards"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/shards/cache"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/tests/docker"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/tests/helper"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/tests/mocks"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/vid"
	jwt_http_router "github.com/SmartEnergyPlatform/jwt-http-router"
	"net/http/httptest"
	"testing"
)

const jwt jwt_http_router.JwtImpersonate = `Bearer eyJhbGciOiJSUzI1NiIsInR5cCIgOiAiSldUIiwia2lkIiA6ICJsWVh1Y1NFMHVQcFpDUHhZX3Q1WEVnMlRsWUoyTVl0TWhwN1hLNThsbmJvIn0.eyJqdGkiOiIwOGM0N2E4OC0yYzc5LTQyMGYtODEwNC02NWJkOWViYmU0MWUiLCJleHAiOjE1NDY1MDcyMzMsIm5iZiI6MCwiaWF0IjoxNTQ2NTA3MTczLCJpc3MiOiJodHRwOi8vbG9jYWxob3N0OjgwMDEvYXV0aC9yZWFsbXMvbWFzdGVyIiwiYXVkIjoiZnJvbnRlbmQiLCJzdWIiOiIzN2MyM2QzMC00YjQ4LTQyMDktOWJkNy0wMzcxZjYyYzJjZmYiLCJ0eXAiOiJCZWFyZXIiLCJhenAiOiJmcm9udGVuZCIsIm5vbmNlIjoiOTJjNDNjOTUtNzViMC00NmNmLTgwYWUtNDVkZDk3M2I0YjdmIiwiYXV0aF90aW1lIjoxNTQ2NTA3MDA5LCJzZXNzaW9uX3N0YXRlIjoiNWRmOTI4ZjQtMDhmMC00ZWI5LTliNjAtM2EwYWUyMmVmYzczIiwiYWNyIjoiMCIsImFsbG93ZWQtb3JpZ2lucyI6WyIqIl0sInJlYWxtX2FjY2VzcyI6eyJyb2xlcyI6WyJjcmVhdGUtcmVhbG0iLCJhZG1pbiIsInVtYV9hdXRob3JpemF0aW9uIl19LCJyZXNvdXJjZV9hY2Nlc3MiOnsibWFzdGVyLXJlYWxtIjp7InJvbGVzIjpbInZpZXctcmVhbG0iLCJ2aWV3LWlkZW50aXR5LXByb3ZpZGVycyIsIm1hbmFnZS1pZGVudGl0eS1wcm92aWRlcnMiLCJpbXBlcnNvbmF0aW9uIiwiY3JlYXRlLWNsaWVudCIsIm1hbmFnZS11c2VycyIsInF1ZXJ5LXJlYWxtcyIsInZpZXctYXV0aG9yaXphdGlvbiIsInF1ZXJ5LWNsaWVudHMiLCJxdWVyeS11c2VycyIsIm1hbmFnZS1ldmVudHMiLCJtYW5hZ2UtcmVhbG0iLCJ2aWV3LWV2ZW50cyIsInZpZXctdXNlcnMiLCJ2aWV3LWNsaWVudHMiLCJtYW5hZ2UtYXV0aG9yaXphdGlvbiIsIm1hbmFnZS1jbGllbnRzIiwicXVlcnktZ3JvdXBzIl19LCJhY2NvdW50Ijp7InJvbGVzIjpbIm1hbmFnZS1hY2NvdW50IiwibWFuYWdlLWFjY291bnQtbGlua3MiLCJ2aWV3LXByb2ZpbGUiXX19LCJyb2xlcyI6WyJhZG1pbiIsImNyZWF0ZS1yZWFsbSIsIm9mZmxpbmVfYWNjZXNzIiwidW1hX2F1dGhvcml6YXRpb24iXSwicHJlZmVycmVkX3VzZXJuYW1lIjoic2VwbCJ9.cSWTHIOHkugQcVNgatbXjvDIP_Ir_QKuUuozbyweh1dJEFsZToTjJ4-5w947bLETmqiNElqXlIV8dT4c9DnPoiXAzsdSotkzKFEYEqRhjYm2obc7Wine1rVwFC4b0Tc5voIzCPNVGFlJDFYWqsPuQYNvAuCIs_A4W86AXWAuxzTyBk5gcRVBLLkFX6GErS2a_4jKd0m26Wd3qoO_j5cl2z2r0AtJ5py4PESiTRLDxEiMoahVQ4coYtX2esWoCRpkSa-beqlD8ffuKaHt95Z8AVcGjBZeSuZpVq6qY6bPBasqVdNkq-CvSnXqWnzNhvq2lUPt58Wp7jeMIJQG4015Zg`

func TestDeploymentStart(t *testing.T) {
	cqrs := mocks.Kafka()

	pgCloser, _, _, pgStr, err := docker.Helper_getPgDependency("vid_relations")
	defer pgCloser()
	if err != nil {
		t.Error(err)
		return
	}

	camundaPgCloser, _, camundaPgIp, _, err := docker.Helper_getPgDependency("camunda")
	defer camundaPgCloser()
	if err != nil {
		t.Error(err)
		return
	}

	camundaCloser, camundaPort, _, err := docker.Helper_getCamundaDependency(camundaPgIp, "5432")
	defer camundaCloser()
	if err != nil {
		t.Error(err)
		return
	}

	config, err := configuration.LoadConfig("../../config.json")
	if err != nil {
		t.Error(err)
		return
	}

	config.WrapperDb = pgStr
	config.ShardingDb = pgStr

	s, err := shards.New(config.ShardingDb, cache.None)
	if err != nil {
		t.Error(err)
		return
	}
	err = s.EnsureShard("http://localhost:" + camundaPort)
	if err != nil {
		t.Error(err)
		return
	}

	v, err := vid.New(config.WrapperDb)
	if err != nil {
		t.Error(err)
		return
	}

	c := camunda.New(v, s)

	e, err := events.New(config, cqrs, v, c)
	if err != nil {
		t.Error(err)
		return
	}

	httpServer := httptest.NewServer(GetRouter(config, c, e))
	defer httpServer.Close()

	//put process
	err = helper.PutProcess(e, "1", "n11", helper.JwtPayload.UserId)
	if err != nil {
		t.Error(err)
		return
	}
	resp, err := jwt.Get(httpServer.URL + "/deployment/1/start")
	if err != nil {
		t.Error(err)
		return
	}
	processinstance := model.ProcessInstance{}
	err = json.NewDecoder(resp.Body).Decode(&processinstance)
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(processinstance)
}

func TestDeploymentStartWithSource(t *testing.T) {
	cqrs := mocks.Kafka()

	pgCloser, _, _, pgStr, err := docker.Helper_getPgDependency("vid_relations")
	defer pgCloser()
	if err != nil {
		t.Error(err)
		return
	}

	camundaPgCloser, _, camundaPgIp, _, err := docker.Helper_getPgDependency("camunda")
	defer camundaPgCloser()
	if err != nil {
		t.Error(err)
		return
	}

	camundaCloser, camundaPort, _, err := docker.Helper_getCamundaDependency(camundaPgIp, "5432")
	defer camundaCloser()
	if err != nil {
		t.Error(err)
		return
	}

	config, err := configuration.LoadConfig("../../config.json")
	if err != nil {
		t.Error(err)
		return
	}

	config.WrapperDb = pgStr
	config.ShardingDb = pgStr

	s, err := shards.New(config.ShardingDb, cache.None)
	if err != nil {
		t.Error(err)
		return
	}
	err = s.EnsureShard("http://localhost:" + camundaPort)
	if err != nil {
		t.Error(err)
		return
	}

	v, err := vid.New(config.WrapperDb)
	if err != nil {
		t.Error(err)
		return
	}

	c := camunda.New(v, s)

	e, err := events.New(config, cqrs, v, c)
	if err != nil {
		t.Error(err)
		return
	}

	httpServer := httptest.NewServer(GetRouter(config, c, e))
	defer httpServer.Close()

	//put process
	err = helper.PutProcessWithSource(e, "1", "n11", helper.JwtPayload.UserId, "")
	if err != nil {
		t.Error(err)
		return
	}

	t.Run("check source = ''", CheckDeploymentList(httpServer.URL, "", 1))
	t.Run("check source = 'sepl'", CheckDeploymentList(httpServer.URL, "sepl", 1))
	t.Run("check source = 'generated'", CheckDeploymentList(httpServer.URL, "generated", 0))

	err = helper.PutProcessWithSource(e, "2", "n2", helper.JwtPayload.UserId, "generated")
	if err != nil {
		t.Error(err)
		return
	}

	t.Run("check source = ''", CheckDeploymentList(httpServer.URL, "", 2))
	t.Run("check source = 'sepl'", CheckDeploymentList(httpServer.URL, "sepl", 1))
	t.Run("check source = 'generated'", CheckDeploymentList(httpServer.URL, "generated", 1))

	err = helper.DeleteProcess(e, "1", helper.JwtPayload.UserId)
	if err != nil {
		t.Error(err)
		return
	}

	t.Run("check source = ''", CheckDeploymentList(httpServer.URL, "", 1))
	t.Run("check source = 'sepl'", CheckDeploymentList(httpServer.URL, "sepl", 0))
	t.Run("check source = 'generated'", CheckDeploymentList(httpServer.URL, "generated", 1))
}

func CheckDeploymentList(url string, source string, expectedCount int) func(t *testing.T) {
	return func(t *testing.T) {
		path := "/deployment"
		if source != "" {
			path = path + "?source=" + source
		}
		list := []interface{}{}
		err := jwt.GetJSON(url+path, &list)
		if err != nil {
			t.Fatal(err)
		}
		if len(list) != expectedCount {
			t.Fatal(len(list), expectedCount, list)
		}
	}
}
