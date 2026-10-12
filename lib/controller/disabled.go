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

package controller

import (
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/camunda"
)

// SetDeploymentDisabled switches the deployment with the given virtual id off or on.
// The caller has to check that userId may access the deployment.
func (this *Controller) SetDeploymentDisabled(userId string, vid string, disabled bool) error {
	if disabled {
		return this.disableDeployment(userId, vid)
	}
	return this.enableDeployment(userId, vid)
}

// The intent is persisted first, so a failure after it leaves the deployment marked and a repeated call completes it.
func (this *Controller) disableDeployment(userId string, vid string) error {
	err := this.vid.SetDisabledIntent(vid)
	if err != nil {
		return err
	}
	deploymentId, exists, err := this.vid.GetDeploymentId(vid)
	if err != nil {
		return err
	}
	if !exists {
		return camunda.UnknownVid
	}
	return this.applyDisabled(deploymentId, userId)
}

// The intent is removed only after activation succeeded, so a failed activation stays disabled.
func (this *Controller) enableDeployment(userId string, vid string) error {
	deploymentId, exists, err := this.vid.GetDeploymentId(vid)
	if err != nil {
		return err
	}
	if !exists {
		return camunda.UnknownVid
	}
	definitions, err := this.camunda.GetRawDefinitionsByDeployment(deploymentId, userId)
	if err != nil {
		return err
	}
	for _, definition := range definitions {
		if definition.Suspended {
			err = this.camunda.SetProcessDefinitionSuspended(definition.Id, userId, false)
			if err != nil {
				return err
			}
		}
	}
	return this.vid.RemoveDisabledIntent(vid)
}

// applyDisabled suspends the definitions of the camunda deployment, then ends its running instances.
// Instances are listed after the suspension so none can start in between.
func (this *Controller) applyDisabled(deploymentId string, userId string) error {
	definitions, err := this.camunda.GetRawDefinitionsByDeployment(deploymentId, userId)
	if err != nil {
		return err
	}
	for _, definition := range definitions {
		if !definition.Suspended {
			err = this.camunda.SetProcessDefinitionSuspended(definition.Id, userId, true)
			if err != nil {
				return err
			}
		}
	}
	instances, err := this.camunda.GetRawInstancesByDeployment(deploymentId, userId)
	if err != nil {
		return err
	}
	for _, instance := range instances {
		err = this.camunda.RemoveProcessInstanceIfExists(instance.Id, userId)
		if err != nil {
			return err
		}
	}
	return nil
}
