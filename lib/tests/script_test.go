/*
 * Copyright 2024 InfAI (CC SES)
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
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/configuration"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/events/messages"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/tests/helper"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/tests/resources"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/tests/server"
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"
)

func TestScriptCleanup(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	wg := sync.WaitGroup{}
	defer wg.Wait()
	defer cancel()

	config, err := configuration.LoadConfig("../../config.json")
	if err != nil {
		t.Error(err)
		return
	}

	config, wrapperUrl, _, e, err := server.CreateTestEnv(ctx, &wg, config)
	if err != nil {
		t.Error(err)
		return
	}

	t.Run("deploy", func(t *testing.T) {
		err = e.HandleDeploymentCreate(helper.JwtPayload.GetUserId(), "test", "test", resources.ScriptTest, helper.SvgExample, "", &messages.IncidentHandling{})
		if err != nil {
			t.Error(err)
			return
		}
	})

	t.Run("start", func(t *testing.T) {
		req, err := http.NewRequest("GET", wrapperUrl+"/deployment/test/start", nil)
		if err != nil {
			t.Error(err)
			return
		}
		req.Header.Set("Authorization", string(helper.Jwt))
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Error(err)
			return
		}
		defer resp.Body.Close()
		temp, _ := io.ReadAll(resp.Body)
		if resp.StatusCode != http.StatusInternalServerError {
			t.Error(string(temp))
			return
		}
		if string(temp) == "EOF" {
			t.Error(string(temp))
			return
		}
		if !strings.Contains(string(temp), "Cannot read property \\\"System\\\" from undefined") {
			t.Error(string(temp))
			return
		}
	})
}
