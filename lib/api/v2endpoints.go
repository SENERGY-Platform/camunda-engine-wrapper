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
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/auth"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/camunda"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/configuration"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/events"
	"github.com/julienschmidt/httprouter"
	"log"
	"net/http"
	"time"

	"io"
)

func init() {
	endpoints = append(endpoints, V2Endpoints)
}

func V2Endpoints(config configuration.Config, router *httprouter.Router, c *camunda.Camunda, e *events.Events, m Metrics) {

	router.GET("/v2/process-definitions/:id/start", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		id := params.ByName("id")

		token, err := auth.GetParsedToken(request)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}

		if err := c.CheckProcessDefinitionAccess(id, token.GetUserId()); err != nil {
			log.Println("WARNING: Access denied for user;", token.GetUserId(), err)
			http.Error(writer, "Access denied", http.StatusUnauthorized)
			return
		}

		inputs := parseQueryParameter(request.URL.Query())

		err = c.StartProcess(id, token.GetUserId(), inputs)
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
		json.NewEncoder(writer).Encode([]string{})
	})

	router.GET("/v2/process-definitions/:id/start/id", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		id := params.ByName("id")

		token, err := auth.GetParsedToken(request)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}

		if err := c.CheckProcessDefinitionAccess(id, token.GetUserId()); err != nil {
			log.Println("WARNING: Access denied for user;", token.GetUserId(), err)
			http.Error(writer, "Access denied", http.StatusUnauthorized)
			return
		}

		inputs := parseQueryParameter(request.URL.Query())

		result, err := c.StartProcessGetId(id, token.GetUserId(), inputs)
		if err != nil {
			log.Println("ERROR: error on process start", err)
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(writer).Encode(result)
	})

	router.GET("/v2/deployments/:id", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		id := params.ByName("id")
		token, err := auth.GetParsedToken(request)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		if err := c.CheckDeploymentAccess(id, token.GetUserId()); err != nil {
			log.Println("WARNING: Access denied for user;", token.GetUserId(), err)
			http.Error(writer, "Access denied", http.StatusUnauthorized)
			return
		}
		result, err := c.GetDeployment(id, token.GetUserId())
		if err != nil {
			log.Println("ERROR: error on getDeployment", err)
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(writer).Encode(result)
	})

	router.GET("/v2/deployments/:id/exists", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		id := params.ByName("id")

		token, err := auth.GetParsedToken(request)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}

		err = c.CheckDeploymentAccess(id, token.GetUserId())
		if err == camunda.UnknownVid || err == camunda.CamundaDeploymentUnknown {
			json.NewEncoder(writer).Encode(false)
			return
		}
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(writer).Encode(true)
	})

	router.GET("/v2/deployments/:id/start", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		id := params.ByName("id")

		token, err := auth.GetParsedToken(request)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}

		if err := c.CheckDeploymentAccess(id, token.GetUserId()); err != nil {
			log.Println("WARNING: Access denied for user;", token.GetUserId(), err)
			http.Error(writer, "Access denied", http.StatusUnauthorized)
			return
		}
		definitions, err := c.GetDefinitionByDeploymentVid(id, token.GetUserId())
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

		result, err := c.StartProcessGetId(definitions[0].Id, token.GetUserId(), inputs)
		if err != nil {
			log.Println("ERROR: error on process start", err)
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(writer).Encode(result)
	})

	router.GET("/v2/deployments/:id/parameter", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		id := params.ByName("id")

		token, err := auth.GetParsedToken(request)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}

		if err := c.CheckDeploymentAccess(id, token.GetUserId()); err != nil {
			log.Println("WARNING: Access denied for user;", token.GetUserId(), err)
			http.Error(writer, "Access denied", http.StatusUnauthorized)
			return
		}
		definitions, err := c.GetDefinitionByDeploymentVid(id, token.GetUserId())
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

		result, err := c.GetProcessParameters(definitions[0].Id, token.GetUserId())
		if err != nil {
			log.Println("ERROR: error on process start", err)
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(writer).Encode(result)
	})

	router.GET("/v2/deployments/:id/definition", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		id := params.ByName("id")

		token, err := auth.GetParsedToken(request)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}

		if err := c.CheckDeploymentAccess(id, token.GetUserId()); err != nil {
			log.Println("WARNING: Access denied for user;", token.GetUserId(), id, err)
			http.Error(writer, "Access denied", http.StatusUnauthorized)
			return
		}
		result, err := c.GetDefinitionByDeploymentVid(id, token.GetUserId())
		if err != nil {
			log.Println("ERROR: error on getDeploymentByDef", err)
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(writer).Encode(result)
	})

	router.GET("/v2/deployments/:id/instances", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		id := params.ByName("id")
		token, err := auth.GetParsedToken(request)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		result, err := c.GetInstancesByDeploymentVid(id, token.GetUserId())
		if err == camunda.UnknownVid {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(writer).Encode(result)
	})

	router.GET("/v2/deployments", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		token, err := auth.GetParsedToken(request)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		result, err := c.GetExtendedDeploymentList(token.GetUserId(), request.URL.Query())
		if err == camunda.UnknownVid {
			log.Println("WARNING: unable to use vid for process; try repeat")
			time.Sleep(1 * time.Second)
			result, err = c.GetExtendedDeploymentList(token.GetUserId(), request.URL.Query())
		}
		if err != nil {
			log.Println("ERROR: error on getDeploymentList", err)
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(writer).Encode(result)
	})

	router.GET("/v2/process-definitions/:id", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		id := params.ByName("id")

		token, err := auth.GetParsedToken(request)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}

		if err := c.CheckProcessDefinitionAccess(id, token.GetUserId()); err != nil {
			log.Println("WARNING: Access denied for user;", token.GetUserId(), err)
			http.Error(writer, "Access denied", http.StatusUnauthorized)
			return
		}
		result, err := c.GetProcessDefinition(id, token.GetUserId())
		if err != nil {
			log.Println("ERROR: error on getProcessDefinition", err)
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(writer).Encode(result)
	})

	router.GET("/v2/process-definitions/:id/diagram", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		id := params.ByName("id")
		token, err := auth.GetParsedToken(request)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		if err := c.CheckProcessDefinitionAccess(id, token.GetUserId()); err != nil {
			log.Println("WARNING: Access denied for user;", token.GetUserId(), err)
			http.Error(writer, "Access denied", http.StatusUnauthorized)
			return
		}
		result, err := c.GetProcessDefinitionDiagram(id, token.GetUserId())
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

	router.GET("/v2/process-instances", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		token, err := auth.GetParsedToken(request)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		result, err := c.GetProcessInstanceList(token.GetUserId())
		if err != nil {
			log.Println("ERROR: error on getProcessInstanceList", err)
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(writer).Encode(result)
	})

	router.GET("/v2/process-instances/count", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		token, err := auth.GetParsedToken(request)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		result, err := c.GetProcessInstanceCount(token.GetUserId())
		if err != nil {
			log.Println("ERROR: error on getProcessInstanceCount", err)
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(writer).Encode(result)
	})

	router.GET("/v2/history/process-instances", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		token, err := auth.GetParsedToken(request)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		query := request.URL.Query()
		if query.Get("with_total") == "true" {
			delete(query, "with_total")
			result, err := c.GetFilteredProcessInstanceHistoryListWithTotal(token.GetUserId(), query)
			if err != nil {
				log.Println("ERROR: error on getFilteredProcessInstanceHistoryList", err)
				http.Error(writer, err.Error(), http.StatusInternalServerError)
				return
			}
			json.NewEncoder(writer).Encode(result)
		} else {
			result, err := c.GetFilteredProcessInstanceHistoryList(token.GetUserId(), query)
			if err != nil {
				log.Println("ERROR: error on getFilteredProcessInstanceHistoryList", err)
				http.Error(writer, err.Error(), http.StatusInternalServerError)
				return
			}
			json.NewEncoder(writer).Encode(result)
		}

	})

	router.DELETE("/v2/history/process-instances/:id", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		id := params.ByName("id")
		token, err := auth.GetParsedToken(request)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		definitionId, err := c.CheckHistoryAccess(id, token.GetUserId())
		if err != nil {
			log.Println("WARNING: Access denied for user;", token.GetUserId(), err)
			http.Error(writer, "Access denied", http.StatusUnauthorized)
			return
		}
		err = e.PublishIncidentDeleteByProcessInstanceEvent(id, definitionId)
		if err != nil {
			log.Println("ERROR: error on PublishIncidentDeleteByProcessInstanceEvent", err)
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		err = c.RemoveProcessInstanceHistory(id, token.GetUserId())
		if err != nil {
			log.Println("ERROR: error on removeProcessInstanceHistory", err)
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(writer).Encode("ok")
	})

	router.DELETE("/v2/process-instances/:id", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		id := params.ByName("id")
		token, err := auth.GetParsedToken(request)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		if err := c.CheckProcessInstanceAccess(id, token.GetUserId()); err != nil {
			log.Println("WARNING: Access denied for user;", token.GetUserId(), err)
			http.Error(writer, "Access denied", http.StatusUnauthorized)
			return
		}
		err = c.RemoveProcessInstance(id, token.GetUserId())
		if err != nil {
			log.Println("ERROR: error on removeProcessInstance", err)
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(writer).Encode("ok")
	})

	router.PUT("/v2/process-instances/:id/variables/:variable_name", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		id := params.ByName("id")
		varName := params.ByName("variable_name")
		var varValue interface{}
		err := json.NewDecoder(request.Body).Decode(&varValue)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		token, err := auth.GetParsedToken(request)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		if err := c.CheckProcessInstanceAccess(id, token.GetUserId()); err != nil {
			log.Println("WARNING: Access denied for user;", token.GetUserId(), err)
			http.Error(writer, "Access denied", http.StatusUnauthorized)
			return
		}
		err = c.SetProcessInstanceVariable(id, token.GetUserId(), varName, varValue)
		if err != nil {
			log.Println("ERROR: error on variable update", err)
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(writer).Encode("ok")
	})

	router.DELETE("/v2/process-instances", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		token, err := auth.GetParsedToken(request)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		ids := []string{}
		err = json.NewDecoder(request.Body).Decode(&ids)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		for _, id := range ids {
			if err := c.CheckProcessInstanceAccess(id, token.GetUserId()); err != nil {
				log.Println("WARNING: Access denied for user;", token.GetUserId(), err)
				http.Error(writer, "Access denied", http.StatusUnauthorized)
				return
			}
			err := c.RemoveProcessInstance(id, token.GetUserId())
			if err != nil {
				log.Println("ERROR: error on removeProcessInstance", err)
				http.Error(writer, err.Error(), http.StatusInternalServerError)
				return
			}
		}
		json.NewEncoder(writer).Encode("ok")
		return
	})

	router.DELETE("/v2/history/process-instances", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		token, err := auth.GetParsedToken(request)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		ids := []string{}
		err = json.NewDecoder(request.Body).Decode(&ids)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		for _, id := range ids {
			definitionId, err := c.CheckHistoryAccess(id, token.GetUserId())
			if err != nil {
				log.Println("WARNING: Access denied for user;", token.GetUserId(), err)
				http.Error(writer, "Access denied", http.StatusUnauthorized)
				return
			}
			err = e.PublishIncidentDeleteByProcessInstanceEvent(id, definitionId)
			if err != nil {
				log.Println("ERROR: error on PublishIncidentDeleteByProcessInstanceEvent", err)
				http.Error(writer, err.Error(), http.StatusInternalServerError)
				return
			}
			err = c.RemoveProcessInstanceHistory(id, token.GetUserId())
			if err != nil {
				log.Println("ERROR: error on removeProcessInstanceHistory", err)
				http.Error(writer, err.Error(), http.StatusInternalServerError)
				return
			}
		}
		json.NewEncoder(writer).Encode("ok")
		return
	})

	router.POST("/v2/event-trigger", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		m.NotifyEventTrigger()

		token, err := auth.GetParsedToken(request)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		body, err := io.ReadAll(request.Body)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		msg := map[string]interface{}{}
		err = json.Unmarshal(body, &msg)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		userId := token.GetUserId()
		if token.IsAdmin() {
			temp, ok := msg["tenantId"]
			if ok {
				userId, ok = temp.(string)
				if !ok {
					http.Error(writer, "expect string in tenantId", http.StatusBadRequest)
					return
				}
			}
		}
		resp, err := c.SendEventTrigger(userId, msg)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Fprint(writer, resp)
		return
	})
}
