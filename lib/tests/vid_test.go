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
	"context"
	"encoding/json"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/api"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/camunda"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/client"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/configuration"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/controller"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/metrics"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/model"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/shards"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/shards/cache"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/tests/docker"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/tests/helper"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/vid"
	"log"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

func TestVid(t *testing.T) {
	wg := &sync.WaitGroup{}
	defer wg.Wait()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pgStr, _, _, err := docker.PostgresWithNetwork(ctx, wg, "vid_relations")
	if err != nil {
		t.Error(err)
		return
	}

	_, camundaPgIp, _, err := docker.PostgresWithNetwork(ctx, wg, "camunda")
	if err != nil {
		t.Error(err)
		return
	}

	camundaUrl, err := docker.Camunda(ctx, wg, camundaPgIp, "5432")
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

	incidentApiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer incidentApiServer.Close()
	config.IncidentApiUrl = incidentApiServer.URL

	s, err := shards.New(config.ShardingDb, cache.None)
	if err != nil {
		t.Error(err)
		return
	}
	err = s.EnsureShard(camundaUrl)
	if err != nil {
		t.Error(err)
		return
	}

	v, err := vid.New(config.WrapperDb)
	if err != nil {
		t.Error(err)
		return
	}

	c := camunda.New(config, v, s, nil)

	ctrl := controller.New(config, c, v, nil)

	httpServer := httptest.NewServer(api.GetRouter(config, c, ctrl, metrics.New()))
	defer httpServer.Close()

	wrapperClient := client.New(httpServer.URL)

	//put process
	err = helper.PutProcess(wrapperClient, "1", "n11", helper.JwtPayload.GetUserId())
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
	resp, err := helper.JwtGet(helper.Jwt, httpServer.URL+"/deployment")
	if err != nil {
		t.Error(err)
		return
	}
	deployments := model.CamundaDeployments{}
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
	err = helper.PutProcess(wrapperClient, "1", "n12", helper.JwtPayload.GetUserId())
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

	resp, err = helper.JwtGet(helper.Jwt, httpServer.URL+"/deployment")
	if err != nil {
		t.Error(err)
		return
	}
	deployments = model.CamundaDeployments{}
	err = json.NewDecoder(resp.Body).Decode(&deployments)
	if err != nil {
		t.Error(err)
		return
	}
	if len(deployments) != 1 || deployments[0].Name != "n12" || deployments[0].Id != "1" {
		t.Errorf("unexpected result %#v:", deployments)
		return
	}

	//delete by vid
	err = helper.DeleteProcess(wrapperClient, "1", helper.JwtPayload.GetUserId())
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

	resp, err = helper.JwtGet(helper.Jwt, httpServer.URL+"/deployment")
	if err != nil {
		t.Error(err)
		return
	}
	deployments = model.CamundaDeployments{}
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
	err = helper.PutProcess(wrapperClient, "v2", "n2", helper.JwtPayload.GetUserId())
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

	resp, err = helper.JwtGet(helper.Jwt, httpServer.URL+"/deployment")
	if err != nil {
		t.Error(err)
		return
	}
	deployments = model.CamundaDeployments{}
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
	err = helper.DeleteProcess(wrapperClient, "v3", helper.JwtPayload.GetUserId())
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

	resp, err = helper.JwtGet(helper.Jwt, httpServer.URL+"/deployment")
	if err != nil {
		t.Error(err)
		return
	}
	deployments = model.CamundaDeployments{}
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
	err = helper.DeleteProcess(wrapperClient, "v4", helper.JwtPayload.GetUserId())
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

	resp, err = helper.JwtGet(helper.Jwt, httpServer.URL+"/deployment")
	if err != nil {
		t.Error(err)
		return
	}
	deployments = model.CamundaDeployments{}
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
