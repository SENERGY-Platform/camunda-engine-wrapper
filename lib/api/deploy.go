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

package api

import (
	"encoding/json"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/auth"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/camunda"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/configuration"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/controller"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/model"
	"net/http"
)

func init() {
	endpoints = append(endpoints, &DeployEndpoints{})
}

type DeployEndpoints struct{}

// Deploy godoc
// @Summary      deploy process
// @Description  deploy process, only admins may access this endpoint
// @Tags         deployment
// @Security Bearer
// @Param        message body model.DeploymentMessage true "deployment"
// @Success      200
// @Failure      400
// @Failure      401
// @Failure      403
// @Failure      404
// @Failure      500
// @Router       /deployment [PUT]
func (this *DeployEndpoints) Deploy(config configuration.Config, router *http.ServeMux, c *camunda.Camunda, e *controller.Controller, m Metrics) {
	router.HandleFunc("PUT /deployment", func(writer http.ResponseWriter, request *http.Request) {
		depl := model.DeploymentMessage{}
		err := json.NewDecoder(request.Body).Decode(&depl)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		token, err := auth.GetParsedToken(request)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		if !token.IsAdmin() {
			http.Error(writer, "only admins may create deployments", http.StatusForbidden)
			return
		}
		err, code := e.Deploy(depl)
		if err != nil {
			http.Error(writer, err.Error(), code)
			return
		}
		writer.WriteHeader(http.StatusOK)
	})
}

// DeleteDeployment godoc
// @Summary      delete deployment
// @Description  delete deployment, only admins may access this endpoint
// @Tags         deployment
// @Security Bearer
// @Param        deplid path string true "deployment id"
// @Param        userid path string true "user id"
// @Success      200
// @Failure      400
// @Failure      401
// @Failure      403
// @Failure      404
// @Failure      500
// @Router       /deployment/{userid}/{deplid} [DELETE]
func (this *DeployEndpoints) DeleteDeployment(config configuration.Config, router *http.ServeMux, c *camunda.Camunda, e *controller.Controller, m Metrics) {
	router.HandleFunc("DELETE /deployment/{userid}/{deplid}", func(writer http.ResponseWriter, request *http.Request) {
		userid := request.PathValue("userid")
		deplid := request.PathValue("deplid")
		token, err := auth.GetParsedToken(request)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		if !token.IsAdmin() {
			http.Error(writer, "only admins may delete deployments", http.StatusForbidden)
			return
		}
		err = e.DeleteDeployment(userid, deplid)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		writer.WriteHeader(http.StatusOK)
	})
}
