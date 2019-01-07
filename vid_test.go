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

package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/SmartEnergyPlatform/jwt-http-router"
	"github.com/ory/dockertest"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
)

const bpmnExample = `<bpmn:definitions xmlns:xsi='http://www.w3.org/2001/XMLSchema-instance' xmlns:bpmn='http://www.omg.org/spec/BPMN/20100524/MODEL' xmlns:bpmndi='http://www.omg.org/spec/BPMN/20100524/DI' xmlns:dc='http://www.omg.org/spec/DD/20100524/DC' xmlns:di='http://www.omg.org/spec/DD/20100524/DI' id='Definitions_1' targetNamespace='http://bpmn.io/schema/bpmn'><bpmn:process id='Process_1' isExecutable='true'><bpmn:startEvent id='StartEvent_1'><bpmn:outgoing>SequenceFlow_02ibfc0</bpmn:outgoing></bpmn:startEvent><bpmn:endEvent id='EndEvent_1728wjv'><bpmn:incoming>SequenceFlow_02ibfc0</bpmn:incoming></bpmn:endEvent><bpmn:sequenceFlow id='SequenceFlow_02ibfc0' sourceRef='StartEvent_1' targetRef='EndEvent_1728wjv'/></bpmn:process><bpmndi:BPMNDiagram id='BPMNDiagram_1'><bpmndi:BPMNPlane id='BPMNPlane_1' bpmnElement='Process_1'><bpmndi:BPMNShape id='_BPMNShape_StartEvent_2' bpmnElement='StartEvent_1'><dc:Bounds x='173' y='102' width='36' height='36'/></bpmndi:BPMNShape><bpmndi:BPMNShape id='EndEvent_1728wjv_di' bpmnElement='EndEvent_1728wjv'><dc:Bounds x='259' y='102' width='36' height='36'/></bpmndi:BPMNShape><bpmndi:BPMNEdge id='SequenceFlow_02ibfc0_di' bpmnElement='SequenceFlow_02ibfc0'><di:waypoint x='209' y='120'></di:waypoint><di:waypoint x='259' y='120'></di:waypoint></bpmndi:BPMNEdge></bpmndi:BPMNPlane></bpmndi:BPMNDiagram></bpmn:definitions>`
const svgExample = `<svg height='48' version='1.1' viewBox='167 96 134 48' width='134' xmlns='http://www.w3.org/2000/svg' xmlns:xlink='http://www.w3.org/1999/xlink'><defs><marker id='sequenceflow-end-white-black-3gh21e50i1p8scvqmvrotmp9p' markerHeight='10' markerWidth='10' orient='auto' refX='11' refY='10' viewBox='0 0 20 20'><path d='M 1 5 L 11 10 L 1 15 Z' style='fill: black; stroke-width: 1px; stroke-linecap: round; stroke-dasharray: 10000, 1; stroke: black;'/></marker></defs><g class='djs-group'><g class='djs-element djs-connection' data-element-id='SequenceFlow_04zz9eb' style='display: block;'><g class='djs-visual'><path d='m  209,120L259,120 ' style='fill: none; stroke-width: 2px; stroke: black; stroke-linejoin: round; marker-end: url(&apos;#sequenceflow-end-white-black-3gh21e50i1p8scvqmvrotmp9p&apos;);'/></g><polyline class='djs-hit' points='209,120 259,120 ' style='fill: none; stroke-opacity: 0; stroke: white; stroke-width: 15px;'/><rect class='djs-outline' height='12' style='fill: none;' width='62' x='203' y='114'/></g></g><g class='djs-group'><g class='djs-element djs-shape' data-element-id='StartEvent_1' style='display: block;' transform='translate(173 102)'><g class='djs-visual'><circle cx='18' cy='18' r='18' style='stroke: black; stroke-width: 2px; fill: white; fill-opacity: 0.95;'/><path d='m 8.459999999999999,11.34 l 0,12.6 l 18.900000000000002,0 l 0,-12.6 z l 9.450000000000001,5.4 l 9.450000000000001,-5.4' style='fill: white; stroke-width: 1px; stroke: black;'/></g><rect class='djs-hit' height='36' style='fill: none; stroke-opacity: 0; stroke: white; stroke-width: 15px;' width='36' x='0' y='0'></rect><rect class='djs-outline' height='48' style='fill: none;' width='48' x='-6' y='-6'></rect></g></g><g class='djs-group'><g class='djs-element djs-shape' data-element-id='EndEvent_056p30q' style='display: block;' transform='translate(259 102)'><g class='djs-visual'><circle cx='18' cy='18' r='18' style='stroke: black; stroke-width: 4px; fill: white; fill-opacity: 0.95;'/></g><rect class='djs-hit' height='36' style='fill: none; stroke-opacity: 0; stroke: white; stroke-width: 15px;' width='36' x='0' y='0'></rect><rect class='djs-outline' height='48' style='fill: none;' width='48' x='-6' y='-6'></rect></g></g></svg>`

const jwt jwt_http_router.JwtImpersonate = `Bearer eyJhbGciOiJSUzI1NiIsInR5cCIgOiAiSldUIiwia2lkIiA6ICJsWVh1Y1NFMHVQcFpDUHhZX3Q1WEVnMlRsWUoyTVl0TWhwN1hLNThsbmJvIn0.eyJqdGkiOiIwOGM0N2E4OC0yYzc5LTQyMGYtODEwNC02NWJkOWViYmU0MWUiLCJleHAiOjE1NDY1MDcyMzMsIm5iZiI6MCwiaWF0IjoxNTQ2NTA3MTczLCJpc3MiOiJodHRwOi8vbG9jYWxob3N0OjgwMDEvYXV0aC9yZWFsbXMvbWFzdGVyIiwiYXVkIjoiZnJvbnRlbmQiLCJzdWIiOiIzN2MyM2QzMC00YjQ4LTQyMDktOWJkNy0wMzcxZjYyYzJjZmYiLCJ0eXAiOiJCZWFyZXIiLCJhenAiOiJmcm9udGVuZCIsIm5vbmNlIjoiOTJjNDNjOTUtNzViMC00NmNmLTgwYWUtNDVkZDk3M2I0YjdmIiwiYXV0aF90aW1lIjoxNTQ2NTA3MDA5LCJzZXNzaW9uX3N0YXRlIjoiNWRmOTI4ZjQtMDhmMC00ZWI5LTliNjAtM2EwYWUyMmVmYzczIiwiYWNyIjoiMCIsImFsbG93ZWQtb3JpZ2lucyI6WyIqIl0sInJlYWxtX2FjY2VzcyI6eyJyb2xlcyI6WyJjcmVhdGUtcmVhbG0iLCJhZG1pbiIsInVtYV9hdXRob3JpemF0aW9uIl19LCJyZXNvdXJjZV9hY2Nlc3MiOnsibWFzdGVyLXJlYWxtIjp7InJvbGVzIjpbInZpZXctcmVhbG0iLCJ2aWV3LWlkZW50aXR5LXByb3ZpZGVycyIsIm1hbmFnZS1pZGVudGl0eS1wcm92aWRlcnMiLCJpbXBlcnNvbmF0aW9uIiwiY3JlYXRlLWNsaWVudCIsIm1hbmFnZS11c2VycyIsInF1ZXJ5LXJlYWxtcyIsInZpZXctYXV0aG9yaXphdGlvbiIsInF1ZXJ5LWNsaWVudHMiLCJxdWVyeS11c2VycyIsIm1hbmFnZS1ldmVudHMiLCJtYW5hZ2UtcmVhbG0iLCJ2aWV3LWV2ZW50cyIsInZpZXctdXNlcnMiLCJ2aWV3LWNsaWVudHMiLCJtYW5hZ2UtYXV0aG9yaXphdGlvbiIsIm1hbmFnZS1jbGllbnRzIiwicXVlcnktZ3JvdXBzIl19LCJhY2NvdW50Ijp7InJvbGVzIjpbIm1hbmFnZS1hY2NvdW50IiwibWFuYWdlLWFjY291bnQtbGlua3MiLCJ2aWV3LXByb2ZpbGUiXX19LCJyb2xlcyI6WyJhZG1pbiIsImNyZWF0ZS1yZWFsbSIsIm9mZmxpbmVfYWNjZXNzIiwidW1hX2F1dGhvcml6YXRpb24iXSwicHJlZmVycmVkX3VzZXJuYW1lIjoic2VwbCJ9.cSWTHIOHkugQcVNgatbXjvDIP_Ir_QKuUuozbyweh1dJEFsZToTjJ4-5w947bLETmqiNElqXlIV8dT4c9DnPoiXAzsdSotkzKFEYEqRhjYm2obc7Wine1rVwFC4b0Tc5voIzCPNVGFlJDFYWqsPuQYNvAuCIs_A4W86AXWAuxzTyBk5gcRVBLLkFX6GErS2a_4jKd0m26Wd3qoO_j5cl2z2r0AtJ5py4PESiTRLDxEiMoahVQ4coYtX2esWoCRpkSa-beqlD8ffuKaHt95Z8AVcGjBZeSuZpVq6qY6bPBasqVdNkq-CvSnXqWnzNhvq2lUPt58Wp7jeMIJQG4015Zg`

var jwtPayload = jwt_http_router.Jwt{}
var _ = jwt_http_router.GetJWTPayload(string(jwt), &jwtPayload)

func TestVid(t *testing.T) {
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

	configLocation := flag.String("config", "config.json", "configuration file")
	flag.Parse()

	err = LoadConfig(*configLocation)
	if err != nil {
		t.Error(err)
		return
	}

	Config.PgConn = pgStr
	Config.ProcessEngineUrl = "http://localhost:" + camundaPort

	httpServer := httptest.NewServer(getRoutes())
	defer httpServer.Close()

	db, err := GetDB()
	if err != nil {
		t.Error(err)
		return
	}

	//put process
	err = testHelper_putProcess("1", "n11", jwtPayload.UserId)
	if err != nil {
		t.Error(err)
		return
	}
	//check relations and process
	byVid, byDeplId, err := getVidRelations(db)
	log.Println(byVid, byDeplId)
	if len(byVid) != 1 || byVid["1"] == "" {
		t.Error("unexpected result:", byVid)
		return
	}
	resp, err := jwt.Get(httpServer.URL + "/deployment")
	if err != nil {
		t.Error(err)
		return
	}
	deployments := Deployments{}
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
	err = testHelper_putProcess("1", "n12", jwtPayload.UserId)
	if err != nil {
		t.Error(err)
		return
	}

	//check relations and process (name is updated; no new processes)
	byVid, byDeplId, err = getVidRelations(db)
	log.Println(byVid, byDeplId)
	if len(byVid) != 1 || byVid["1"] == "" {
		t.Error("unexpected result:", byVid)
		return
	}

	resp, err = jwt.Get(httpServer.URL + "/deployment")
	if err != nil {
		t.Error(err)
		return
	}
	deployments = Deployments{}
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
	err = testHelper_deleteProcess("1")
	if err != nil {
		t.Error(err)
		return
	}

	//check relations and process (removed)
	byVid, byDeplId, err = getVidRelations(db)
	log.Println(byVid, byDeplId)
	if len(byVid) != 0 {
		t.Error("unexpected result:", byVid)
		return
	}

	resp, err = jwt.Get(httpServer.URL + "/deployment")
	if err != nil {
		t.Error(err)
		return
	}
	deployments = Deployments{}
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
	err = saveVidRelation("v2", "d2")
	if err != nil {
		t.Error(err)
		return
	}

	//put process matching relation by "event"
	err = testHelper_putProcess("v2", "n2", jwtPayload.UserId)
	if err != nil {
		t.Error(err)
		return
	}

	//check relations and process (update relation and add process)
	byVid, byDeplId, err = getVidRelations(db)
	log.Println(byVid, byDeplId)
	if len(byVid) != 1 || byVid["v2"] == "d2" || byVid["v2"] == "" {
		t.Error("unexpected result:", byVid)
		return
	}

	resp, err = jwt.Get(httpServer.URL + "/deployment")
	if err != nil {
		t.Error(err)
		return
	}
	deployments = Deployments{}
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
	err = saveVidRelation("v3", "d3")
	if err != nil {
		t.Error(err)
		return
	}

	//delete added relation by "event"
	err = testHelper_deleteProcess("v3")
	if err != nil {
		t.Error(err)
		return
	}

	//check relations and process (removed)
	byVid, byDeplId, err = getVidRelations(db)
	log.Println(byVid, byDeplId)
	if len(byVid) != 1 || byVid["v2"] == "d2" || byVid["v2"] == "" {
		t.Error("unexpected result:", byVid)
		return
	}

	resp, err = jwt.Get(httpServer.URL + "/deployment")
	if err != nil {
		t.Error(err)
		return
	}
	deployments = Deployments{}
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
	err = testHelper_deleteProcess("v4")
	if err != nil {
		t.Error(err)
		return
	}

	//check relations and process (no change)
	byVid, byDeplId, err = getVidRelations(db)
	log.Println(byVid, byDeplId)
	if len(byVid) != 1 || byVid["v2"] == "d2" || byVid["v2"] == "" {
		t.Error("unexpected result:", byVid)
		return
	}

	resp, err = jwt.Get(httpServer.URL + "/deployment")
	if err != nil {
		t.Error(err)
		return
	}
	deployments = Deployments{}
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

func testHelper_getPgDependency(dbName string) (closer func(), hostPort string, ipAddress string, pgStr string, err error) {
	pool, err := dockertest.NewPool("")
	if err != nil {
		return func() {}, "", "", "", err
	}
	log.Println("start postgres db")
	pg, err := pool.Run("postgres", "latest", []string{"POSTGRES_DB=" + dbName, "POSTGRES_PASSWORD=pw", "POSTGRES_USER=usr"})
	if err != nil {
		return func() {}, "", "", "", err
	}
	hostPort = pg.GetPort("5432/tcp")
	pgStr = fmt.Sprintf("postgres://usr:pw@localhost:%s/%s?sslmode=disable", hostPort, dbName)
	err = pool.Retry(func() error {
		log.Println("try pg connection...")
		var err error
		db, err = sql.Open("postgres", pgStr)
		if err != nil {
			return err
		}
		return db.Ping()
	})
	return func() { pg.Close() }, hostPort, pg.Container.NetworkSettings.IPAddress, pgStr, err
}

func testHelper_getCamundaDependency(pgIp string, pgPort string) (closer func(), hostPort string, ipAddress string, err error) {
	pool, err := dockertest.NewPool("")
	if err != nil {
		return func() {}, "", "", err
	}
	log.Println("start process engine")
	camunda, err := pool.Run("fgseitsrancher.wifa.intern.uni-leipzig.de:5000/process-engine", "unstable", []string{
		"DB_PASSWORD=pw",
		"DB_URL=jdbc:postgresql://" + pgIp + ":" + pgPort + "/camunda",
		"DB_PORT=" + pgPort,
		"DB_NAME=camunda",
		"DB_HOST=" + pgIp,
		"DB_DRIVER=org.postgresql.Driver",
		"DB_USERNAME=usr",
		"DATABASE=postgres",
	})
	if err != nil {
		return func() {}, "", "", err
	}
	hostPort = camunda.GetPort("8080/tcp")
	err = pool.Retry(func() error {
		log.Println("try camunda connection...")
		resp, err := http.Get("http://localhost:" + hostPort + "/engine-rest/metrics")
		if err != nil {
			return err
		}
		if resp.StatusCode != 200 {
			log.Println("unexpectet response code", resp.StatusCode, resp.Status)
			return errors.New("unexpectet response code: " + resp.Status)
		}
		return nil
	})
	return func() { camunda.Close() }, hostPort, camunda.Container.NetworkSettings.IPAddress, err
}

func testHelper_putProcess(vid string, name string, owner string) error {
	return handleDeploymentCreate(DeploymentCommand{
		Id:            vid,
		Command:       "PUT",
		Owner:         owner,
		DeploymentXml: bpmnExample,
		Deployment: DeploymentRequest{
			Svg: svgExample,
			Process: AbstractProcess{
				Name: name,
			},
		},
	})
}

func testHelper_deleteProcess(vid string) error {
	return handleDeploymentDelete(vid)
}
