/*
 * Copyright 2021 InfAI (CC SES)
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

package cleanup

import "github.com/SmartEnergyPlatform/util/http/request"

//returns all process deployments without replacing the deployment id with the virtual id
func getDeploymentListAllRaw(shard string) (result Deployments, err error) {
	path := shard + "/engine-rest/deployment"
	err = request.Get(path, &result)
	return
}

type Deployment struct {
	Id             string `json:"id"`
	Name           string `json:"name"`
	Source         string `json:"source"`
	DeploymentTime string `json:"deploymentTime"`
	TenantId       string `json:"tenantId"`
}

type Deployments = []Deployment
