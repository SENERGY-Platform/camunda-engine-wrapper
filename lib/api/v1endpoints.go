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

package api

import (
	"encoding/json"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/auth"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/camunda"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/configuration"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/events"
	"github.com/julienschmidt/httprouter"
	"log"
	"net/http"
	"net/url"
	"time"

	"io"
)

func init() {
	endpoints = append(endpoints, V1Endpoints)
}

func V1Endpoints(config configuration.Config, router *httprouter.Router, c *camunda.Camunda, e *events.Events, m Metrics) {

	router.GET("/", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		writer.Header().Set("Content-Type", "application/json")
		json.NewEncoder(writer).Encode(map[string]string{"status": "OK"})
	})

	router.GET("/process-definition/:id/start", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
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
		writer.Header().Set("Content-Type", "application/json")
		json.NewEncoder(writer).Encode([]string{})
	})

	router.GET("/process-definition/:id/start/id", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
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
		writer.Header().Set("Content-Type", "application/json")
		json.NewEncoder(writer).Encode(result)
	})

	router.GET("/deployment/:id", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		//"/engine-rest/deployment/" + id
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
		writer.Header().Set("Content-Type", "application/json")
		json.NewEncoder(writer).Encode(result)
	})

	router.GET("/deployment/:id/exists", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		//"/engine-rest/deployment/" + id
		id := params.ByName("id")

		token, err := auth.GetParsedToken(request)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}

		err = c.CheckDeploymentAccess(id, token.GetUserId())
		if err == camunda.UnknownVid || err == camunda.CamundaDeploymentUnknown {
			writer.Header().Set("Content-Type", "application/json")
			json.NewEncoder(writer).Encode(false)
			return
		}
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		writer.Header().Set("Content-Type", "application/json")
		json.NewEncoder(writer).Encode(true)
	})

	router.GET("/deployment/:id/start", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
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
		writer.Header().Set("Content-Type", "application/json")
		json.NewEncoder(writer).Encode(result)
	})

	router.GET("/deployment/:id/parameter", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
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
		writer.Header().Set("Content-Type", "application/json")
		json.NewEncoder(writer).Encode(result)
	})

	router.GET("/deployment/:id/definition", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		//"/engine-rest/process-definition?deploymentId=
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
		writer.Header().Set("Content-Type", "application/json")
		json.NewEncoder(writer).Encode(result)
	})

	router.GET("/deployment", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
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
		writer.Header().Set("Content-Type", "application/json")
		json.NewEncoder(writer).Encode(result)
	})

	router.GET("/process-definition/:id", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		//"/engine-rest/process-definition/" + processDefinitionId
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
		writer.Header().Set("Content-Type", "application/json")
		json.NewEncoder(writer).Encode(result)
	})

	router.GET("/process-definition/:id/diagram", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		// "/engine-rest/process-definition/" + processDefinitionId + "/diagram"
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

	router.GET("/process-instance", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		//"/engine-rest/process-instance"

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
		writer.Header().Set("Content-Type", "application/json")
		json.NewEncoder(writer).Encode(result)
	})

	router.GET("/process-instances/count", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		//"/engine-rest/process-instance/count"

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
		writer.Header().Set("Content-Type", "application/json")
		json.NewEncoder(writer).Encode(result)
	})

	router.GET("/history/process-instance", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		//"/engine-rest/history/process-instance"

		token, err := auth.GetParsedToken(request)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}

		result, err := c.GetProcessInstanceHistoryList(token.GetUserId())
		if err != nil {
			log.Println("ERROR: error on getProcessInstanceHistoryList", err)
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		writer.Header().Set("Content-Type", "application/json")
		json.NewEncoder(writer).Encode(result)
	})

	router.GET("/history/filtered/process-instance", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		//"/engine-rest/history/process-instance"

		token, err := auth.GetParsedToken(request)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}

		result, err := c.GetFilteredProcessInstanceHistoryList(token.GetUserId(), request.URL.Query())
		if err != nil {
			log.Println("ERROR: error on getFilteredProcessInstanceHistoryList", err)
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		writer.Header().Set("Content-Type", "application/json")
		json.NewEncoder(writer).Encode(result)
	})

	router.GET("/history/finished/process-instance", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		//"/engine-rest/history/process-instance"
		token, err := auth.GetParsedToken(request)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		result, err := c.GetProcessInstanceHistoryListFinished(token.GetUserId())
		if err != nil {
			log.Println("ERROR: error on getProcessInstanceHistoryListFinished", err)
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		writer.Header().Set("Content-Type", "application/json")
		json.NewEncoder(writer).Encode(result)
	})

	router.GET("/history/finished/process-instance/:searchtype/:searchvalue/:limit/:offset/:sortby/:sortdirection", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		searchvalue := params.ByName("searchvalue")
		searchtype := params.ByName("searchtype")
		limit := params.ByName("limit")
		offset := params.ByName("offset")
		sortby := params.ByName("sortby")
		sortdirection := params.ByName("sortdirection")

		token, err := auth.GetParsedToken(request)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}

		result, err := c.GetProcessInstanceHistoryListWithTotal(token.GetUserId(), searchtype, searchvalue, limit, offset, sortby, sortdirection, true)
		if err != nil {
			log.Println("ERROR: error on getProcessInstanceHistoryListWithTotal", err)
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		writer.Header().Set("Content-Type", "application/json")
		json.NewEncoder(writer).Encode(result)
	})

	router.GET("/history/unfinished/process-instance/:searchtype/:searchvalue/:limit/:offset/:sortby/:sortdirection", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		searchvalue := params.ByName("searchvalue")
		searchtype := params.ByName("searchtype")
		limit := params.ByName("limit")
		offset := params.ByName("offset")
		sortby := params.ByName("sortby")
		sortdirection := params.ByName("sortdirection")

		token, err := auth.GetParsedToken(request)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}

		result, err := c.GetProcessInstanceHistoryListWithTotal(token.GetUserId(), searchtype, searchvalue, limit, offset, sortby, sortdirection, false)
		if err != nil {
			log.Println("ERROR: error on getProcessInstanceHistoryListWithTotal", err)
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		writer.Header().Set("Content-Type", "application/json")
		json.NewEncoder(writer).Encode(result)
	})

	router.GET("/history/unfinished/process-instance", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		//"/engine-rest/history/process-instance"
		token, err := auth.GetParsedToken(request)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		result, err := c.GetProcessInstanceHistoryListUnfinished(token.GetUserId())
		if err != nil {
			log.Println("ERROR: error on getProcessInstanceHistoryListUnfinished", err)
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		writer.Header().Set("Content-Type", "application/json")
		json.NewEncoder(writer).Encode(result)
	})

	router.GET("/history/process-definition/:id/process-instance", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		//"/engine-rest/history/process-instance?processDefinitionId="
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
		result, err := c.GetProcessInstanceHistoryByProcessDefinition(id, token.GetUserId())
		if err != nil {
			log.Println("ERROR: error on processinstanceHistoryByDefinition", err)
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		writer.Header().Set("Content-Type", "application/json")
		json.NewEncoder(writer).Encode(result)
	})

	router.GET("/history/process-definition/:id/process-instance/finished", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		//"/engine-rest/history/process-instance?processDefinitionId="
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
		result, err := c.GetProcessInstanceHistoryByProcessDefinitionFinished(id, token.GetUserId())
		if err != nil {
			log.Println("ERROR: error on processinstanceHistoryByDefinition", err)
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		writer.Header().Set("Content-Type", "application/json")
		json.NewEncoder(writer).Encode(result)
	})

	router.GET("/history/process-definition/:id/process-instance/unfinished", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		//"/engine-rest/history/process-instance?processDefinitionId="
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
		result, err := c.GetProcessInstanceHistoryByProcessDefinitionUnfinished(id, token.GetUserId())
		if err != nil {
			log.Println("ERROR: error on processinstanceHistoryByDefinition", err)
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		writer.Header().Set("Content-Type", "application/json")
		json.NewEncoder(writer).Encode(result)
	})

	router.DELETE("/history/process-instance/:id", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		//DELETE "/engine-rest/history/process-instance/" + processInstanceId
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
		writer.Header().Set("Content-Type", "application/json")
		json.NewEncoder(writer).Encode("ok")
	})

	router.DELETE("/process-instance/:id", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		//DELETE "/engine-rest/process-instance/" + processInstanceId
		id := params.ByName("id")

		token, err := auth.GetParsedToken(request)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}

		if err, code := c.CheckProcessInstanceAccess(id, token.GetUserId()); err != nil {
			http.Error(writer, err.Error(), code)
			return
		}
		err = c.RemoveProcessInstance(id, token.GetUserId())
		if err != nil {
			log.Println("ERROR: error on removeProcessInstance", err)
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		writer.Header().Set("Content-Type", "application/json")
		json.NewEncoder(writer).Encode("ok")
	})
}

func parseQueryParameter(query url.Values) (result map[string]interface{}) {
	if len(query) == 0 {
		return map[string]interface{}{}
	}
	result = map[string]interface{}{}
	for key, _ := range query {
		var val interface{}
		temp := query.Get(key)
		err := json.Unmarshal([]byte(temp), &val)
		if err != nil {
			val = temp
		}
		result[key] = val
	}
	return result
}
