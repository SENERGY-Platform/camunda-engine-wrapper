/*
 * Copyright 2025 InfAI (CC SES)
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

package model

import "github.com/SENERGY-Platform/models/go/models"

type Deployment struct {
	Id               string            `json:"id"`
	Name             string            `json:"name"`
	Diagram          Diagram           `json:"diagram"`
	IncidentHandling *IncidentHandling `json:"incident_handling,omitempty"`
}

type IncidentHandling = models.IncidentHandling

type Diagram = models.Diagram

type DeploymentMessage struct {
	Deployment
	UserId string `json:"user_id"`
	Source string `json:"source"` //optional
}
