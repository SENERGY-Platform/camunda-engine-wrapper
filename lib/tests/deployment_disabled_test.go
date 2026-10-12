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
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"sync"
	"testing"

	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/auth"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/camunda"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/client"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/configuration"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/model"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/tests/helper"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/tests/resources"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/tests/server"
	gojwt "github.com/golang-jwt/jwt"
)

const disabledTestMessageName = "disable_test_message"

func TestDeploymentDisabled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	wg := sync.WaitGroup{}
	defer wg.Wait()
	defer cancel()

	config, err := configuration.LoadConfig("../../config.json")
	if err != nil {
		t.Error(err)
		return
	}

	incidentApi := &disabledTestIncidentApi{suspendedAtRegistration: map[string]bool{}}
	config, wrapperUrl, shard, err := server.CreateTestEnvWithIncidentApi(ctx, &wg, config, incidentApi)
	if err != nil {
		t.Error(err)
		return
	}
	incidentApi.setShard(shard)

	wrapperClient := client.New(wrapperUrl)
	otherUserToken, err := disabledTestToken("disable-test-other-user")
	if err != nil {
		t.Error(err)
		return
	}

	const deplId = "disable_test"

	t.Run("deploy", func(t *testing.T) {
		disabledTestDeploy(wrapperClient, deplId)(t)
		incidentApi.expectSuspendedAtRegistration(t, disabledTestDefinitionId(t, wrapperUrl, deplId), false)
	})

	t.Run("start an instance that runs for an hour", func(t *testing.T) {
		_, err, code := wrapperClient.StartDeployment(helper.Jwt, deplId, client.StartOptions{})
		if err != nil {
			t.Error(code, err)
		}
		disabledTestExpectRunningInstances(t, wrapperUrl, deplId, 1)
	})

	t.Run("another user gets 401 and the deployment stays active", func(t *testing.T) {
		code, body := disabledTestPut(t, wrapperUrl, otherUserToken, deplId, `{"disabled": true}`)
		if code != http.StatusUnauthorized {
			t.Error(code, body)
		}
		disabledTestExpectSuspended(t, wrapperUrl, deplId, false)
	})

	t.Run("unknown deployment gets 401", func(t *testing.T) {
		code, body := disabledTestPut(t, wrapperUrl, helper.Jwt, "unknown_deployment", `{"disabled": true}`)
		if code != http.StatusUnauthorized {
			t.Error(code, body)
		}
	})

	t.Run("bad body gets 400 and the deployment stays active", func(t *testing.T) {
		for _, body := range []string{``, `{}`, `{"disabled": null}`, `{"disabled": "true"}`, `not json`} {
			code, resp := disabledTestPut(t, wrapperUrl, helper.Jwt, deplId, body)
			if code != http.StatusBadRequest {
				t.Errorf("body %q: %v %v", body, code, resp)
			}
		}
		disabledTestExpectSuspended(t, wrapperUrl, deplId, false)
	})

	t.Run("disable answers 200 with the state", disabledTestSet(wrapperUrl, deplId, true))
	t.Run("disable again is idempotent", disabledTestSet(wrapperUrl, deplId, true))

	t.Run("definition is suspended", func(t *testing.T) {
		disabledTestExpectSuspended(t, wrapperUrl, deplId, true)
	})

	t.Run("running instance was removed", func(t *testing.T) {
		disabledTestExpectRunningInstances(t, wrapperUrl, deplId, 0)
	})

	t.Run("list shows disabled", disabledTestExpectListed(wrapperClient, deplId, true))

	t.Run("message start event starts nothing", func(t *testing.T) {
		disabledTestTriggerMessage(t, wrapperUrl)
		disabledTestExpectRunningInstances(t, wrapperUrl, deplId, 0)
	})

	t.Run("start endpoints answer 409", func(t *testing.T) {
		definitionId := disabledTestDefinitionId(t, wrapperUrl, deplId)
		for _, path := range []string{
			"/v2/deployments/" + deplId + "/start",
			"/deployment/" + deplId + "/start",
			"/v2/process-definitions/" + definitionId + "/start",
			"/v2/process-definitions/" + definitionId + "/start/id",
			"/process-definition/" + definitionId + "/start",
			"/process-definition/" + definitionId + "/start/id",
		} {
			resp, err := helper.JwtGet(helper.Jwt, wrapperUrl+path)
			if err != nil {
				t.Error(path, err)
				continue
			}
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			if resp.StatusCode != http.StatusConflict || strings.TrimSpace(string(body)) != "deployment is disabled" {
				t.Error(path, resp.StatusCode, string(body))
			}
		}
		disabledTestExpectRunningInstances(t, wrapperUrl, deplId, 0)
	})

	t.Run("redeploy with the same id stays disabled", func(t *testing.T) {
		before := disabledTestDefinitionId(t, wrapperUrl, deplId)
		disabledTestDeploy(wrapperClient, deplId)(t)
		after := disabledTestDefinitionId(t, wrapperUrl, deplId)
		if before == after {
			t.Error("expected a new process definition after redeploy", before)
		}
		incidentApi.expectSuspendedAtRegistration(t, after, true)
		disabledTestExpectSuspended(t, wrapperUrl, deplId, true)
		disabledTestExpectListed(wrapperClient, deplId, true)(t)
		_, err, code := wrapperClient.StartDeployment(helper.Jwt, deplId, client.StartOptions{})
		if code != http.StatusConflict {
			t.Error(code, err)
		}
	})

	t.Run("redeploy that falls back to the placeholder process stays disabled", func(t *testing.T) {
		// the message name is already subscribed by deplId, so camunda rejects this deployment and the wrapper deploys a placeholder
		const dupId = "disable_test_duplicate_message"
		disabledTestDeploy(wrapperClient, dupId)(t)
		disabledTestSet(wrapperUrl, dupId, true)(t)
		disabledTestDeploy(wrapperClient, dupId)(t)
		definitions := disabledTestDefinitions(t, wrapperUrl, dupId)
		if len(definitions) != 1 || !strings.HasPrefix(definitions[0].Key, "id_") {
			t.Fatalf("expected one placeholder definition, got %#v", definitions)
		}
		incidentApi.expectSuspendedAtRegistration(t, definitions[0].Id, true)
		disabledTestExpectSuspended(t, wrapperUrl, dupId, true)
	})

	t.Run("enable answers 200 with the state", disabledTestSet(wrapperUrl, deplId, false))
	t.Run("enable again is idempotent", disabledTestSet(wrapperUrl, deplId, false))

	t.Run("list shows enabled", disabledTestExpectListed(wrapperClient, deplId, false))

	t.Run("manual start works after enable", func(t *testing.T) {
		disabledTestExpectSuspended(t, wrapperUrl, deplId, false)
		_, err, code := wrapperClient.StartDeployment(helper.Jwt, deplId, client.StartOptions{})
		if err != nil {
			t.Error(code, err)
		}
		disabledTestExpectRunningInstances(t, wrapperUrl, deplId, 1)
	})

	t.Run("message start works after enable", func(t *testing.T) {
		disabledTestTriggerMessage(t, wrapperUrl)
		disabledTestExpectRunningInstances(t, wrapperUrl, deplId, 2)
	})

	t.Run("delete removes the disabled state of a redeployed id", func(t *testing.T) {
		disabledTestSet(wrapperUrl, deplId, true)(t)
		err := helper.DeleteProcess(wrapperClient, deplId, helper.JwtPayload.GetUserId())
		if err != nil {
			t.Fatal(err)
		}
		disabledTestDeploy(wrapperClient, deplId)(t)
		disabledTestExpectSuspended(t, wrapperUrl, deplId, false)
		disabledTestExpectListed(wrapperClient, deplId, false)(t)
		_, err, code := wrapperClient.StartDeployment(helper.Jwt, deplId, client.StartOptions{})
		if err != nil {
			t.Error(code, err)
		}
	})
}

// disabledTestIncidentApi records whether a definition was already suspended when the wrapper registered its
// incident handler, which happens after the camunda deployment and before the vid relation is saved.
type disabledTestIncidentApi struct {
	mux                     sync.Mutex
	shard                   string
	suspendedAtRegistration map[string]bool
}

func (this *disabledTestIncidentApi) setShard(shard string) {
	this.mux.Lock()
	defer this.mux.Unlock()
	this.shard = shard
}

func (this *disabledTestIncidentApi) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPut || request.URL.Path != "/on-incident-handler" {
		return
	}
	msg := struct {
		ProcessDefinitionId string `json:"process_definition_id"`
	}{}
	err := json.NewDecoder(request.Body).Decode(&msg)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}
	this.mux.Lock()
	shard := this.shard
	this.mux.Unlock()
	definition := model.ProcessDefinition{}
	err = camunda.Get(shard+"/engine-rest/process-definition/"+url.PathEscape(msg.ProcessDefinitionId), &definition)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}
	this.mux.Lock()
	defer this.mux.Unlock()
	this.suspendedAtRegistration[msg.ProcessDefinitionId] = definition.Suspended
}

func (this *disabledTestIncidentApi) expectSuspendedAtRegistration(t *testing.T, definitionId string, expected bool) {
	t.Helper()
	this.mux.Lock()
	defer this.mux.Unlock()
	suspended, ok := this.suspendedAtRegistration[definitionId]
	if !ok {
		t.Errorf("no incident handler registered for %v", definitionId)
		return
	}
	if suspended != expected {
		t.Errorf("definition %v: suspended=%v at incident handler registration, expected %v", definitionId, suspended, expected)
	}
}

// disabledTestToken builds a token for another user; the wrapper does not verify token signatures, so any key works.
func disabledTestToken(userId string) (string, error) {
	token, err := gojwt.NewWithClaims(gojwt.SigningMethodHS256, &auth.Token{Sub: userId}).SignedString([]byte("test"))
	return "Bearer " + token, err
}

func disabledTestDeploy(c *client.Client, id string) func(t *testing.T) {
	return func(t *testing.T) {
		err, code := c.Deploy(client.InternalAdminToken, client.DeploymentMessage{
			Deployment: client.Deployment{
				Id:   id,
				Name: id,
				Diagram: client.Diagram{
					XmlDeployed: resources.MessageStartLong,
					Svg:         helper.SvgExample,
				},
				IncidentHandling: &client.IncidentHandling{},
			},
			UserId: helper.JwtPayload.GetUserId(),
		})
		if err != nil {
			t.Fatal(code, err)
		}
	}
}

func disabledTestPut(t *testing.T, wrapperUrl string, token string, id string, body string) (code int, respBody string) {
	t.Helper()
	req, err := http.NewRequest(http.MethodPut, wrapperUrl+"/v2/deployments/"+url.PathEscape(id)+"/disabled", strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Authorization", token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	temp, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, string(temp)
}

func disabledTestSet(wrapperUrl string, id string, disabled bool) func(t *testing.T) {
	return func(t *testing.T) {
		reqBody, _ := json.Marshal(model.DeploymentDisabled{Disabled: disabled})
		code, body := disabledTestPut(t, wrapperUrl, helper.Jwt, id, string(reqBody))
		if code != http.StatusOK {
			t.Fatal(code, body)
		}
		result := map[string]interface{}{}
		err := json.Unmarshal([]byte(body), &result)
		if err != nil {
			t.Fatal(err, body)
		}
		if len(result) != 1 || result["disabled"] != disabled {
			t.Fatal(body)
		}
	}
}

func disabledTestDefinitions(t *testing.T, wrapperUrl string, id string) (result model.ProcessDefinitions) {
	t.Helper()
	resp, err := helper.JwtGet(helper.Jwt, wrapperUrl+"/v2/deployments/"+url.PathEscape(id)+"/definition")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		temp, _ := io.ReadAll(resp.Body)
		t.Fatal(resp.StatusCode, string(temp))
	}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		t.Fatal(err)
	}
	return result
}

func disabledTestDefinitionId(t *testing.T, wrapperUrl string, id string) string {
	t.Helper()
	definitions := disabledTestDefinitions(t, wrapperUrl, id)
	if len(definitions) != 1 {
		t.Fatalf("expected one definition, got %#v", definitions)
	}
	return definitions[0].Id
}

func disabledTestExpectSuspended(t *testing.T, wrapperUrl string, id string, suspended bool) {
	t.Helper()
	definitions := disabledTestDefinitions(t, wrapperUrl, id)
	if len(definitions) == 0 {
		t.Error("no definitions found for", id)
	}
	for _, definition := range definitions {
		if definition.Suspended != suspended {
			t.Errorf("definition %v of %v: suspended=%v, expected %v", definition.Id, id, definition.Suspended, suspended)
		}
	}
}

func disabledTestExpectRunningInstances(t *testing.T, wrapperUrl string, id string, expected int) {
	t.Helper()
	resp, err := helper.JwtGet(helper.Jwt, wrapperUrl+"/v2/deployments/"+url.PathEscape(id)+"/instances")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	temp, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		t.Fatal(resp.StatusCode, string(temp))
	}
	instances := model.ProcessInstances{}
	err = json.Unmarshal(temp, &instances)
	if err != nil {
		t.Fatal(err, string(temp))
	}
	if len(instances) != expected {
		t.Errorf("expected %v running instances of %v, got %v: %v", expected, id, len(instances), string(temp))
	}
}

func disabledTestExpectListed(c *client.Client, id string, disabled bool) func(t *testing.T) {
	return func(t *testing.T) {
		list, err, code := c.ListDeployments(helper.Jwt, client.DeploymentListOptions{})
		if err != nil {
			t.Fatal(code, err)
		}
		i := slices.IndexFunc(list, func(d model.ExtendedDeployment) bool { return d.Id == id })
		if i < 0 {
			t.Fatalf("deployment %v not listed: %#v", id, list)
		}
		if list[i].Error != "" {
			t.Error(list[i].Error)
		}
		if list[i].Disabled != disabled {
			t.Errorf("listed disabled=%v, expected %v", list[i].Disabled, disabled)
		}
	}
}

// disabledTestTriggerMessage correlates to all subscriptions, which camunda answers with success even if none match.
func disabledTestTriggerMessage(t *testing.T, wrapperUrl string) {
	t.Helper()
	body, _ := json.Marshal(map[string]interface{}{"messageName": disabledTestMessageName, "all": true})
	req, err := http.NewRequest(http.MethodPost, wrapperUrl+"/v2/event-trigger", strings.NewReader(string(body)))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Authorization", helper.Jwt)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	temp, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		t.Fatal(resp.StatusCode, string(temp))
	}
}
