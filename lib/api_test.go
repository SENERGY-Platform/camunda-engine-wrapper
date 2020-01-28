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
	"net/http/httptest"
	"testing"
)

func TestDeploymentStart(t *testing.T) {
	cqrs = SilentKafkaMock{}

	pgCloser, _, _, pgStr, err := testHelper_getPgDependency("vid_relations")
	defer pgCloser()
	if err != nil {
		t.Error(err)
		return
	}

	camundaPgCloser, _, camundaPgIp, _, err := testHelper_getPgDependency("camunda")
	defer camundaPgCloser()
	if err != nil {
		t.Error(err)
		return
	}

	camundaCloser, camundaPort, _, err := testHelper_getCamundaDependency(camundaPgIp, "5432")
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
	Config.ProcessEngineUrl = "http://localhost:" + camundaPort

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
