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
	"encoding/json"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/cache"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/shards"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/tests/docker"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/tests/mocks"
	"net/http/httptest"
	"testing"
)

func TestDeploymentStart(t *testing.T) {
	cqrs = mocks.Kafka()

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

	err = LoadConfig("../config.json")
	if err != nil {
		t.Error(err)
		return
	}

	Config.PgConn = pgStr
	s, err := shards.New(Config.PgConn, cache.None)
	if err != nil {
		t.Error(err)
		return
	}
	err = s.EnsureShard("http://localhost:" + camundaPort)
	if err != nil {
		t.Error(err)
		return
	}

	httpServer := httptest.NewServer(getRoutes())
	defer httpServer.Close()

	//put process
	err = testHelper_putProcess("1", "n11", jwtPayload.UserId)
	if err != nil {
		t.Error(err)
		return
	}
	resp, err := jwt.Get(httpServer.URL + "/deployment/1/start")
	if err != nil {
		t.Error(err)
		return
	}
	processinstance := ProcessInstance{}
	err = json.NewDecoder(resp.Body).Decode(&processinstance)
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(processinstance)
}

func TestDeploymentStartWithSource(t *testing.T) {
	cqrs = mocks.Kafka()

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

	err = LoadConfig("../config.json")
	if err != nil {
		t.Error(err)
		return
	}

	Config.PgConn = pgStr
	s, err := shards.New(Config.PgConn, cache.None)
	if err != nil {
		t.Error(err)
		return
	}
	err = s.EnsureShard("http://localhost:" + camundaPort)
	if err != nil {
		t.Error(err)
		return
	}

	httpServer := httptest.NewServer(getRoutes())
	defer httpServer.Close()

	//put process
	err = testHelper_putProcessWithSource("1", "n11", jwtPayload.UserId, "")
	if err != nil {
		t.Error(err)
		return
	}

	t.Run("check source = ''", CheckDeploymentList(httpServer.URL, "", 1))
	t.Run("check source = 'sepl'", CheckDeploymentList(httpServer.URL, "sepl", 1))
	t.Run("check source = 'generated'", CheckDeploymentList(httpServer.URL, "generated", 0))

	err = testHelper_putProcessWithSource("2", "n2", jwtPayload.UserId, "generated")
	if err != nil {
		t.Error(err)
		return
	}

	t.Run("check source = ''", CheckDeploymentList(httpServer.URL, "", 2))
	t.Run("check source = 'sepl'", CheckDeploymentList(httpServer.URL, "sepl", 1))
	t.Run("check source = 'generated'", CheckDeploymentList(httpServer.URL, "generated", 1))

	err = testHelper_deleteProcess("1", jwtPayload.UserId)
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
