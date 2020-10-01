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

package tests

import (
	"encoding/json"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/api"
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
	"log"
	"net/http/httptest"
	"testing"
)

func TestVid(t *testing.T) {
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

	httpServer := httptest.NewServer(api.GetRouter(config, c, e))
	defer httpServer.Close()

	//put process
	err = helper.PutProcess(e, "1", "n11", helper.JwtPayload.UserId)
	if err != nil {
		t.Error(err)
		return
	}
	//check relations and process
	byVid, byDeplId, err := v.GetRelations()
	log.Println(byVid, byDeplId)
	if len(byVid) != 1 || byVid["1"] == "" {
		t.Error("unexpected result:", byVid)
		return
	}
	resp, err := helper.Jwt.Get(httpServer.URL + "/deployment")
	if err != nil {
		t.Error(err)
		return
	}
	deployments := model.Deployments{}
	err = json.NewDecoder(resp.Body).Decode(&deployments)
	if err != nil {
		t.Error(err)
		return
	}
	if len(deployments) != 1 || deployments[0].Name != "n11" || deployments[0].Id != "1" {
		t.Error("unexpected result:", deployments)
		return
	}

	//overwrite name
	err = helper.PutProcess(e, "1", "n12", helper.JwtPayload.UserId)
	if err != nil {
		t.Error(err)
		return
	}

	//check relations and process (name is updated; no new processes)
	byVid, byDeplId, err = v.GetRelations()
	log.Println(byVid, byDeplId)
	if len(byVid) != 1 || byVid["1"] == "" {
		t.Error("unexpected result:", byVid)
		return
	}

	resp, err = helper.Jwt.Get(httpServer.URL + "/deployment")
	if err != nil {
		t.Error(err)
		return
	}
	deployments = model.Deployments{}
	err = json.NewDecoder(resp.Body).Decode(&deployments)
	if err != nil {
		t.Error(err)
		return
	}
	if len(deployments) != 1 || deployments[0].Name != "n12" || deployments[0].Id != "1" {
		t.Error("unexpected result:", deployments)
		return
	}

	//delete by vid
	err = helper.DeleteProcess(e, "1", helper.JwtPayload.UserId)
	if err != nil {
		t.Error(err)
		return
	}

	//check relations and process (removed)
	byVid, byDeplId, err = v.GetRelations()
	log.Println(byVid, byDeplId)
	if len(byVid) != 0 {
		t.Error("unexpected result:", byVid)
		return
	}

	resp, err = helper.Jwt.Get(httpServer.URL + "/deployment")
	if err != nil {
		t.Error(err)
		return
	}
	deployments = model.Deployments{}
	err = json.NewDecoder(resp.Body).Decode(&deployments)
	if err != nil {
		t.Error(err)
		return
	}
	if len(deployments) != 0 {
		t.Error("unexpected result:", deployments)
		return
	}

	//manually add relation without process in camunda
	err = v.SaveVidRelation("v2", "d2")
	if err != nil {
		t.Error(err)
		return
	}

	//put process matching relation by "event"
	err = helper.PutProcess(e, "v2", "n2", helper.JwtPayload.UserId)
	if err != nil {
		t.Error(err)
		return
	}

	//check relations and process (update relation and add process)
	byVid, byDeplId, err = v.GetRelations()
	log.Println(byVid, byDeplId)
	if len(byVid) != 1 || byVid["v2"] == "d2" || byVid["v2"] == "" {
		t.Error("unexpected result:", byVid)
		return
	}

	resp, err = helper.Jwt.Get(httpServer.URL + "/deployment")
	if err != nil {
		t.Error(err)
		return
	}
	deployments = model.Deployments{}
	err = json.NewDecoder(resp.Body).Decode(&deployments)
	if err != nil {
		t.Error(err)
		return
	}
	if len(deployments) != 1 || deployments[0].Name != "n2" || deployments[0].Id != "v2" {
		t.Error("unexpected result:", deployments)
		return
	}

	//manually add relation without process in camunda
	err = v.SaveVidRelation("v3", "d3")
	if err != nil {
		t.Error(err)
		return
	}

	//delete added relation by "event"
	err = helper.DeleteProcess(e, "v3", helper.JwtPayload.UserId)
	if err != nil {
		t.Error(err)
		return
	}

	//check relations and process (removed)
	byVid, byDeplId, err = v.GetRelations()
	log.Println(byVid, byDeplId)
	if len(byVid) != 1 || byVid["v2"] == "d2" || byVid["v2"] == "" {
		t.Error("unexpected result:", byVid)
		return
	}

	resp, err = helper.Jwt.Get(httpServer.URL + "/deployment")
	if err != nil {
		t.Error(err)
		return
	}
	deployments = model.Deployments{}
	err = json.NewDecoder(resp.Body).Decode(&deployments)
	if err != nil {
		t.Error(err)
		return
	}
	if len(deployments) != 1 || deployments[0].Name != "n2" || deployments[0].Id != "v2" {
		t.Error("unexpected result:", deployments)
		return
	}

	//delete not existing relation (vid) by "event"
	err = helper.DeleteProcess(e, "v4", helper.JwtPayload.UserId)
	if err != nil {
		t.Error(err)
		return
	}

	//check relations and process (no change)
	byVid, byDeplId, err = v.GetRelations()
	log.Println(byVid, byDeplId)
	if len(byVid) != 1 || byVid["v2"] == "d2" || byVid["v2"] == "" {
		t.Error("unexpected result:", byVid)
		return
	}

	resp, err = helper.Jwt.Get(httpServer.URL + "/deployment")
	if err != nil {
		t.Error(err)
		return
	}
	deployments = model.Deployments{}
	err = json.NewDecoder(resp.Body).Decode(&deployments)
	if err != nil {
		t.Error(err)
		return
	}
	if len(deployments) != 1 || deployments[0].Name != "n2" || deployments[0].Id != "v2" {
		t.Error("unexpected result:", deployments)
		return
	}
}
