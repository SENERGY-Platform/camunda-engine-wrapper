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
	"fmt"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/camunda"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/configuration"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/events"
	"log"
	"net/http"
	"time"

	"io"

	"github.com/SmartEnergyPlatform/jwt-http-router"
	"github.com/SmartEnergyPlatform/util/http/response"
)

func init() {
	endpoints = append(endpoints, V2Endpoints)
}

func V2Endpoints(config configuration.Config, router *jwt_http_router.Router, c *camunda.Camunda, e *events.Events) {

	router.GET("/v2/process-definitions/:id/start", func(writer http.ResponseWriter, request *http.Request, params jwt_http_router.Params, jwt jwt_http_router.Jwt) {
		id := params.ByName("id")
		if err := c.CheckProcessDefinitionAccess(id, jwt.UserId); err != nil {
			log.Println("WARNING: Access denied for user;", jwt.UserId, err)
			http.Error(writer, "Access denied", http.StatusUnauthorized)
			return
		}

		inputs := parseQueryParameter(request.URL.Query())

		err := c.StartProcess(id, jwt.UserId, inputs)
		if err != nil {
			log.Println("ERROR: error on process start", err)
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		//old version returned list of process-variables from history
		// variable history is deprecated
		// keep empty list as output
		// 		to keep signature of endpoint
		// 		and ensure that no other services throw errors because of this change
		response.To(writer).Json([]string{})
	})

	router.GET("/v2/process-definitions/:id/start/id", func(writer http.ResponseWriter, request *http.Request, params jwt_http_router.Params, jwt jwt_http_router.Jwt) {
		id := params.ByName("id")
		if err := c.CheckProcessDefinitionAccess(id, jwt.UserId); err != nil {
			log.Println("WARNING: Access denied for user;", jwt.UserId, err)
			http.Error(writer, "Access denied", http.StatusUnauthorized)
			return
		}

		inputs := parseQueryParameter(request.URL.Query())

		result, err := c.StartProcessGetId(id, jwt.UserId, inputs)
		if err != nil {
			log.Println("ERROR: error on process start", err)
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		response.To(writer).Json(result)
	})

	router.GET("/v2/deployments/:id", func(writer http.ResponseWriter, request *http.Request, params jwt_http_router.Params, jwt jwt_http_router.Jwt) {
		id := params.ByName("id")
		if err := c.CheckDeploymentAccess(id, jwt.UserId); err != nil {
			log.Println("WARNING: Access denied for user;", jwt.UserId, err)
			http.Error(writer, "Access denied", http.StatusUnauthorized)
			return
		}
		result, err := c.GetDeployment(id, jwt.UserId)
		if err != nil {
			log.Println("ERROR: error on getDeployment", err)
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		response.To(writer).Json(result)
	})

	router.GET("/v2/deployments/:id/exists", func(writer http.ResponseWriter, request *http.Request, params jwt_http_router.Params, jwt jwt_http_router.Jwt) {
		id := params.ByName("id")
		err := c.CheckDeploymentAccess(id, jwt.UserId)
		if err == camunda.UnknownVid || err == camunda.CamundaDeploymentUnknown {
			response.To(writer).Json(false)
			return
		}
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		response.To(writer).Json(true)
	})

	router.GET("/v2/deployments/:id/start", func(writer http.ResponseWriter, request *http.Request, params jwt_http_router.Params, jwt jwt_http_router.Jwt) {
		id := params.ByName("id")
		if err := c.CheckDeploymentAccess(id, jwt.UserId); err != nil {
			log.Println("WARNING: Access denied for user;", jwt.UserId, err)
			http.Error(writer, "Access denied", http.StatusUnauthorized)
			return
		}
		definitions, err := c.GetDefinitionByDeploymentVid(id, jwt.UserId)
		if err != nil {
			log.Println("ERROR: error on getDeploymentByDef", err)
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		if len(definitions) == 0 {
			log.Println("ERROR: no definition for deployment found", err)
			http.Error(writer, "no definition for deployment found", http.StatusInternalServerError)
			return
		}

		inputs := parseQueryParameter(request.URL.Query())

		result, err := c.StartProcessGetId(definitions[0].Id, jwt.UserId, inputs)
		if err != nil {
			log.Println("ERROR: error on process start", err)
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		response.To(writer).Json(result)
	})

	router.GET("/v2/deployments/:id/parameter", func(writer http.ResponseWriter, request *http.Request, params jwt_http_router.Params, jwt jwt_http_router.Jwt) {
		id := params.ByName("id")
		if err := c.CheckDeploymentAccess(id, jwt.UserId); err != nil {
			log.Println("WARNING: Access denied for user;", jwt.UserId, err)
			http.Error(writer, "Access denied", http.StatusUnauthorized)
			return
		}
		definitions, err := c.GetDefinitionByDeploymentVid(id, jwt.UserId)
		if err != nil {
			log.Println("ERROR: error on getDeploymentByDef", err)
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		if len(definitions) == 0 {
			log.Println("ERROR: no definition for deployment found", err)
			http.Error(writer, "no definition for deployment found", http.StatusInternalServerError)
			return
		}

		result, err := c.GetProcessParameters(definitions[0].Id, jwt.UserId)
		if err != nil {
			log.Println("ERROR: error on process start", err)
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		response.To(writer).Json(result)
	})

	router.GET("/v2/deployments/:id/definition", func(writer http.ResponseWriter, request *http.Request, params jwt_http_router.Params, jwt jwt_http_router.Jwt) {
		id := params.ByName("id")
		if err := c.CheckDeploymentAccess(id, jwt.UserId); err != nil {
			log.Println("WARNING: Access denied for user;", jwt.UserId, id, err)
			http.Error(writer, "Access denied", http.StatusUnauthorized)
			return
		}
		result, err := c.GetDefinitionByDeploymentVid(id, jwt.UserId)
		if err != nil {
			log.Println("ERROR: error on getDeploymentByDef", err)
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		response.To(writer).Json(result)
	})

	router.GET("/v2/deployments", func(writer http.ResponseWriter, request *http.Request, params jwt_http_router.Params, jwt jwt_http_router.Jwt) {
		result, err := c.GetExtendedDeploymentList(jwt.UserId, request.URL.Query())
		if err == camunda.UnknownVid {
			log.Println("WARNING: unable to use vid for process; try repeat")
			time.Sleep(1 * time.Second)
			result, err = c.GetExtendedDeploymentList(jwt.UserId, request.URL.Query())
		}
		if err != nil {
			log.Println("ERROR: error on getDeploymentList", err)
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		response.To(writer).Json(result)
	})

	router.GET("/v2/process-definitions/:id", func(writer http.ResponseWriter, request *http.Request, params jwt_http_router.Params, jwt jwt_http_router.Jwt) {
		id := params.ByName("id")
		if err := c.CheckProcessDefinitionAccess(id, jwt.UserId); err != nil {
			log.Println("WARNING: Access denied for user;", jwt.UserId, err)
			http.Error(writer, "Access denied", http.StatusUnauthorized)
			return
		}
		result, err := c.GetProcessDefinition(id, jwt.UserId)
		if err != nil {
			log.Println("ERROR: error on getProcessDefinition", err)
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		response.To(writer).Json(result)
	})

	router.GET("/v2/process-definitions/:id/diagram", func(writer http.ResponseWriter, request *http.Request, params jwt_http_router.Params, jwt jwt_http_router.Jwt) {
		id := params.ByName("id")
		if err := c.CheckProcessDefinitionAccess(id, jwt.UserId); err != nil {
			log.Println("WARNING: Access denied for user;", jwt.UserId, err)
			http.Error(writer, "Access denied", http.StatusUnauthorized)
			return
		}
		result, err := c.GetProcessDefinitionDiagram(id, jwt.UserId)
		if err != nil {
			log.Println("ERROR: error on getProcessDefinitionDiagram", err)
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		//copy image data
		_, err = io.Copy(writer, result.Body)
		if err != nil {
			log.Println("ERROR: error on diagram response copy", err)
			http.Error(writer, err.Error(), http.StatusNotImplemented)
			return
		}
		//copy image content type
		for key, _ := range result.Header {
			writer.Header().Set(key, result.Header.Get(key))
		}
	})

	router.GET("/v2/process-instances", func(writer http.ResponseWriter, request *http.Request, params jwt_http_router.Params, jwt jwt_http_router.Jwt) {
		result, err := c.GetProcessInstanceList(jwt.UserId)
		if err != nil {
			log.Println("ERROR: error on getProcessInstanceList", err)
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		response.To(writer).Json(result)
	})

	router.GET("/v2/process-instances/count", func(writer http.ResponseWriter, request *http.Request, params jwt_http_router.Params, jwt jwt_http_router.Jwt) {
		result, err := c.GetProcessInstanceCount(jwt.UserId)
		if err != nil {
			log.Println("ERROR: error on getProcessInstanceCount", err)
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		response.To(writer).Json(result)
	})

	router.GET("/v2/history/process-instances", func(writer http.ResponseWriter, request *http.Request, params jwt_http_router.Params, jwt jwt_http_router.Jwt) {
		result, err := c.GetFilteredProcessInstanceHistoryList(jwt.UserId, request.URL.Query())
		if err != nil {
			log.Println("ERROR: error on getFilteredProcessInstanceHistoryList", err)
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		response.To(writer).Json(result)
	})

	router.DELETE("/v2/history/process-instances/:id", func(writer http.ResponseWriter, request *http.Request, params jwt_http_router.Params, jwt jwt_http_router.Jwt) {
		id := params.ByName("id")
		definitionId, err := c.CheckHistoryAccess(id, jwt.UserId)
		if err != nil {
			log.Println("WARNING: Access denied for user;", jwt.UserId, err)
			http.Error(writer, "Access denied", http.StatusUnauthorized)
			return
		}
		err = e.PublishIncidentDeleteByProcessInstanceEvent(id, definitionId)
		if err != nil {
			log.Println("ERROR: error on PublishIncidentDeleteByProcessInstanceEvent", err)
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		err = c.RemoveProcessInstanceHistory(id, jwt.UserId)
		if err != nil {
			log.Println("ERROR: error on removeProcessInstanceHistory", err)
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		response.To(writer).Text("ok")
	})

	router.DELETE("/v2/process-instances/:id", func(writer http.ResponseWriter, request *http.Request, params jwt_http_router.Params, jwt jwt_http_router.Jwt) {
		id := params.ByName("id")
		if err := c.CheckProcessInstanceAccess(id, jwt.UserId); err != nil {
			log.Println("WARNING: Access denied for user;", jwt.UserId, err)
			http.Error(writer, "Access denied", http.StatusUnauthorized)
			return
		}
		err := c.RemoveProcessInstance(id, jwt.UserId)
		if err != nil {
			log.Println("ERROR: error on removeProcessInstance", err)
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		response.To(writer).Text("ok")
	})

	router.DELETE("/v2/process-instances", func(writer http.ResponseWriter, request *http.Request, params jwt_http_router.Params, jwt jwt_http_router.Jwt) {
		ids := []string{}
		err := json.NewDecoder(request.Body).Decode(&ids)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		for _, id := range ids {
			if err := c.CheckProcessInstanceAccess(id, jwt.UserId); err != nil {
				log.Println("WARNING: Access denied for user;", jwt.UserId, err)
				http.Error(writer, "Access denied", http.StatusUnauthorized)
				return
			}
			err := c.RemoveProcessInstance(id, jwt.UserId)
			if err != nil {
				log.Println("ERROR: error on removeProcessInstance", err)
				http.Error(writer, err.Error(), http.StatusInternalServerError)
				return
			}
		}
		response.To(writer).Text("ok")
		return
	})

	router.DELETE("/v2/history/process-instances", func(writer http.ResponseWriter, request *http.Request, params jwt_http_router.Params, jwt jwt_http_router.Jwt) {
		ids := []string{}
		err := json.NewDecoder(request.Body).Decode(&ids)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		for _, id := range ids {
			definitionId, err := c.CheckHistoryAccess(id, jwt.UserId)
			if err != nil {
				log.Println("WARNING: Access denied for user;", jwt.UserId, err)
				http.Error(writer, "Access denied", http.StatusUnauthorized)
				return
			}
			err = e.PublishIncidentDeleteByProcessInstanceEvent(id, definitionId)
			if err != nil {
				log.Println("ERROR: error on PublishIncidentDeleteByProcessInstanceEvent", err)
				http.Error(writer, err.Error(), http.StatusInternalServerError)
				return
			}
			err = c.RemoveProcessInstanceHistory(id, jwt.UserId)
			if err != nil {
				log.Println("ERROR: error on removeProcessInstanceHistory", err)
				http.Error(writer, err.Error(), http.StatusInternalServerError)
				return
			}
		}
		response.To(writer).Text("ok")
		return
	})

	router.POST("/v2/event-trigger", func(writer http.ResponseWriter, request *http.Request, params jwt_http_router.Params, jwt jwt_http_router.Jwt) {
		body, err := io.ReadAll(request.Body)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		response, err := c.SendEventTrigger(jwt.UserId, body)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Fprint(writer, response)
		return
	})
}
