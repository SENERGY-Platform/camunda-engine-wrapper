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
	"errors"
	"flag"
	"fmt"
	"github.com/SmartEnergyPlatform/amqp-wrapper-lib"
	"github.com/SmartEnergyPlatform/jwt-http-router"
	"github.com/SmartEnergyPlatform/util/http/response"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

/*
	TestCases:
		1. process, relation, metadata		000		remains
		2. process, relation, !metadata		001		delete
		3. process, !relation, metadata		010		fixed
		4. process, !relation, !metadata	011		delete
		5. !process, relation, metadata		100		fixed
		6. !process, relation, !metadata	101		delete
		7. !process, !relation, metadata	110		fixed
*/

func TestMigration(t *testing.T) {
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

	configLocation := flag.String("config", "../config.json", "configuration file")
	flag.Parse()

	err = LoadConfig(*configLocation)
	if err != nil {
		t.Error(err)
		return
	}

	deplMock := []DeploymentMetadata{}
	deplMockServer := httptest.NewServer(testHelper_getProcessDeploymentMock(&deplMock))
	defer deplMockServer.Close()

	Config.MigrateProcessDeploymentUrl = deplMockServer.URL
	Config.MigrateJwt = string(jwt)
	Config.PgConn = pgStr
	Config.ProcessEngineUrl = "http://localhost:" + camundaPort

	httpServer := httptest.NewServer(getRoutes())
	defer httpServer.Close()

	//mock amqp
	amqp = &AmqpMock{PublishHandler: func(resource string, payload []byte) error {
		if resource != Config.AmqpDeploymentTopic {
			return errors.New("unexpected resource: " + resource)
		}
		command := DeploymentCommand{} //DeploymentCommand{Id: id, Command: "DELETE"}
		err := json.Unmarshal(payload, &command)
		if err != nil {
			return err
		}
		if command.Command != "DELETE" {
			return errors.New("unexpected command: " + command.Command)
		}
		return testHelper_deleteProcess(command.Id)
	}}

	//case 000
	err = testHelper_putProcess("000", "n000", jwtPayload.UserId)
	if err != nil {
		t.Error(err)
		return
	}
	deplMock = append(deplMock, DeploymentMetadata{Process: "000", Abstract: AbstractProcess{Name: "n000", Xml: bpmnExample}, Owner: jwtPayload.UserId})

	//case 001
	err = testHelper_putProcess("001", "n001", jwtPayload.UserId)
	if err != nil {
		t.Error(err)
		return
	}

	//case 010
	deplId_010, err := DeployProcess("n010", bpmnExample, svgExample, jwtPayload.UserId)
	if err != nil {
		t.Error(err)
		return
	}
	deplMock = append(deplMock, DeploymentMetadata{Process: deplId_010, Abstract: AbstractProcess{Name: "n010", Xml: bpmnExample}, Owner: jwtPayload.UserId})

	//case 011
	_, err = DeployProcess("n011", bpmnExample, svgExample, jwtPayload.UserId)
	if err != nil {
		t.Error(err)
		return
	}

	//case 100a
	err = saveVidRelation("100a", "100a")
	if err != nil {
		t.Error(err)
		return
	}
	deplMock = append(deplMock, DeploymentMetadata{Process: "100a", Abstract: AbstractProcess{Name: "n100a", Xml: bpmnExample}, Owner: jwtPayload.UserId})

	//case 100b
	err = saveVidRelation("100b", "100depl")
	if err != nil {
		t.Error(err)
		return
	}
	deplMock = append(deplMock, DeploymentMetadata{Process: "100b", Abstract: AbstractProcess{Name: "n100b", Xml: bpmnExample}, Owner: jwtPayload.UserId})

	//case 101
	err = saveVidRelation("101", "101")
	if err != nil {
		t.Error(err)
		return
	}

	//case 110
	deplMock = append(deplMock, DeploymentMetadata{Process: "110", Abstract: AbstractProcess{Name: "n110", Xml: bpmnExample}, Owner: jwtPayload.UserId})

	err = Migrate()
	if err != nil {
		t.Error(err)
		return
	}

	byVid, byDeplId, err := getVidRelations(db)
	if err != nil {
		t.Error(err)
		return
	}

	deployments, err := getDeploymentListAllRaw()
	if err != nil {
		t.Error(err)
		return
	}
	deploymentIndex := map[string]Deployment{}
	for _, depl := range deployments {
		deploymentIndex[depl.Id] = depl
	}

	if len(byVid) != 5 {
		t.Error("unexpected byVid length: ", deployments)
		return
	}

	if len(byDeplId) != 5 {
		t.Error("unexpected byDeplId length: ", deployments)
		return
	}

	if len(deployments) != 5 {
		t.Error("unexpected deployments length: ", deployments)
		return
	}

	if byVid["000"] == "000" || byVid["000"] == "" {
		t.Error("unexpected byVid: ", byVid["000"])
		return
	}
	if deploymentIndex[byVid["000"]].Name != "n000" {
		t.Error("unexpected deployment: ", deploymentIndex[byVid["000"]])
	}

	if byVid[deplId_010] != deplId_010 {
		t.Error("unexpected byVid: ", byVid["010"], "|", deplId_010)
		return
	}
	if deploymentIndex[byVid[deplId_010]].Name != "n010" {
		t.Error("unexpected deployment: ", deploymentIndex[byVid[deplId_010]])
	}

	if byVid["100a"] == "" || byVid["100a"] == "100a" {
		t.Error("unexpected byVid: ", byVid["100a"])
		return
	}
	if deploymentIndex[byVid["100a"]].Name != "n100a" {
		t.Error("unexpected deployment: ", deploymentIndex[byVid["100a"]])
	}

	if byVid["100b"] == "" || byVid["100b"] == "100depl" {
		t.Error("unexpected byVid: ", byVid["100b"])
		return
	}
	if deploymentIndex[byVid["100b"]].Name != "n100b" {
		t.Error("unexpected deployment: ", deploymentIndex[byVid["100b"]])
	}

	if byVid["110"] == "" || byVid["110"] == "110" {
		t.Error("unexpected byVid: ", byVid["110"])
		return
	}
	if deploymentIndex[byVid["110"]].Name != "n110" {
		t.Error("unexpected deployment: ", deploymentIndex[byVid["110"]])
		return
	}

	deplInfos, err := getExtendedDeploymentList(jwtPayload.UserId, url.Values{})
	if err != nil {
		t.Error(err)
		return
	}
	if len(deplInfos) != 5 {
		t.Error("unexpected result", len(deplInfos), deplInfos)
	}
	for index, deplInfo := range deplInfos {
		if deplInfo.Deployment.Id == "000" || deplInfo.Deployment.Id == "001" {
			b, _ := json.Marshal(deplInfo)
			fmt.Println(string(b))
			if deplInfo.Id == "" || deplInfo.Name == "" {
				t.Error("unexpected deployment result", index, deplInfo.Deployment)
				return
			}
			if deplInfo.Diagram != svgExample {
				t.Error("unexpected digramm result", index, deplInfo.Diagram, deplInfo)
				return
			}
		}
	}
}

func testHelper_getProcessDeploymentMock(deployments *[]DeploymentMetadata) http.Handler {
	router := jwt_http_router.New(jwt_http_router.JwtConfig{
		ForceUser: true,
		ForceAuth: true,
		PubRsa:    "",
	})

	router.GET("/deployment", func(res http.ResponseWriter, r *http.Request, ps jwt_http_router.Params, jwt jwt_http_router.Jwt) {
		response.To(res).Json(*deployments)
	})
	return router
}

type AmqpMock struct {
	PublishHandler func(resource string, payload []byte) error
}

func (*AmqpMock) Consume(qname string, resource string, worker amqp_wrapper_lib.ConsumerFunc) (err error) {
	panic("implement me")
}

func (this *AmqpMock) Publish(resource string, payload []byte) error {
	return this.PublishHandler(resource, payload)
}

func (*AmqpMock) Close() {
	panic("implement me")
}
