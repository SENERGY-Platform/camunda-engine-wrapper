/*
 * Copyright 2023 InfAI (CC SES)
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
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/camunda/model"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/client"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/configuration"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/tests/resources"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/tests/server"
	"sync"
	"testing"
)

func TestFilteredParams(t *testing.T) {
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

	wrapperClient := client.New(wrapperUrl)

	deploymentId := "withInput"
	t.Run("deploy", testDeployProcessWithInput(wrapperClient, deploymentId, resources.FormFieldTest))

	t.Run("check", checkProcessParameterDeclaration(wrapperUrl, deploymentId, map[string]model.Variable{
		"bar": {
			Value:     "42",
			Type:      "String",
			ValueInfo: map[string]interface{}{},
		},
	}))
}
