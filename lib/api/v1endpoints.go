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
	"log"
	"net/http"
	"net/url"
	"time"

	"io"
)

func init() {
	endpoints = append(endpoints, &V1Endpoints{})
}

type V1Endpoints struct{}

func (this *V1Endpoints) HealthCheck(config configuration.Config, router *http.ServeMux, c *camunda.Camunda, e *events.Events, m Metrics) {
	router.HandleFunc("GET /", func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "application/json")
		json.NewEncoder(writer).Encode(map[string]string{"status": "OK"})
	})
}

func (this *V1Endpoints) StartProcessDefinition(config configuration.Config, router *http.ServeMux, c *camunda.Camunda, e *events.Events, m Metrics) {
	router.HandleFunc("GET /process-definition/{id}/start", func(writer http.ResponseWriter, request *http.Request) {
		id := request.PathValue("id")

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
}

func (this *V1Endpoints) StartProcessDefinitionAndGetInstanceId(config configuration.Config, router *http.ServeMux, c *camunda.Camunda, e *events.Events, m Metrics) {
	router.HandleFunc("GET /process-definition/{id}/start/id", func(writer http.ResponseWriter, request *http.Request) {
		id := request.PathValue("id")

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
}

func (this *V1Endpoints) GetDeployment(config configuration.Config, router *http.ServeMux, c *camunda.Camunda, e *events.Events, m Metrics) {
	router.HandleFunc("GET /deployment/{id}", func(writer http.ResponseWriter, request *http.Request) {
		//"/engine-rest/deployment/" + id
		id := request.PathValue("id")

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
}

func (this *V1Endpoints) DeploymentExists(config configuration.Config, router *http.ServeMux, c *camunda.Camunda, e *events.Events, m Metrics) {
	router.HandleFunc("GET /deployment/{id}/exists", func(writer http.ResponseWriter, request *http.Request) {
		//"/engine-rest/deployment/" + id
		id := request.PathValue("id")

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
}

func (this *V1Endpoints) StartDeployment(config configuration.Config, router *http.ServeMux, c *camunda.Camunda, e *events.Events, m Metrics) {
	router.HandleFunc("GET /deployment/{id}/start", func(writer http.ResponseWriter, request *http.Request) {
		id := request.PathValue("id")

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
}

func (this *V1Endpoints) GetDeploymentParameters(config configuration.Config, router *http.ServeMux, c *camunda.Camunda, e *events.Events, m Metrics) {
	router.HandleFunc("GET /deployment/{id}/parameter", func(writer http.ResponseWriter, request *http.Request) {
		id := request.PathValue("id")

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
}

func (this *V1Endpoints) GetDeploymentDefinition(config configuration.Config, router *http.ServeMux, c *camunda.Camunda, e *events.Events, m Metrics) {
	router.HandleFunc("GET /deployment/{id}/definition", func(writer http.ResponseWriter, request *http.Request) {
		//"/engine-rest/process-definition?deploymentId=
		id := request.PathValue("id")

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
}

func (this *V1Endpoints) GetDeployments(config configuration.Config, router *http.ServeMux, c *camunda.Camunda, e *events.Events, m Metrics) {
	router.HandleFunc("GET /deployment", func(writer http.ResponseWriter, request *http.Request) {
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
}

func (this *V1Endpoints) GetProcessDefinition(config configuration.Config, router *http.ServeMux, c *camunda.Camunda, e *events.Events, m Metrics) {
	router.HandleFunc("GET /process-definition/{id}", func(writer http.ResponseWriter, request *http.Request) {
		//"/engine-rest/process-definition/" + processDefinitionId
		id := request.PathValue("id")

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
}

func (this *V1Endpoints) GetProcessDefinitionDiagram(config configuration.Config, router *http.ServeMux, c *camunda.Camunda, e *events.Events, m Metrics) {
	router.HandleFunc("GET /process-definition/{id}/diagram", func(writer http.ResponseWriter, request *http.Request) {
		// "/engine-rest/process-definition/" + processDefinitionId + "/diagram"
		id := request.PathValue("id")

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
}

func (this *V1Endpoints) GetProcessInstances(config configuration.Config, router *http.ServeMux, c *camunda.Camunda, e *events.Events, m Metrics) {
	router.HandleFunc("GET /process-instance", func(writer http.ResponseWriter, request *http.Request) {
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
}

func (this *V1Endpoints) GetProcessInstanceCount(config configuration.Config, router *http.ServeMux, c *camunda.Camunda, e *events.Events, m Metrics) {
	router.HandleFunc("GET /process-instances/count", func(writer http.ResponseWriter, request *http.Request) {
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
}

func (this *V1Endpoints) GetHistoricProcessInstances(config configuration.Config, router *http.ServeMux, c *camunda.Camunda, e *events.Events, m Metrics) {
	router.HandleFunc("GET /history/process-instance", func(writer http.ResponseWriter, request *http.Request) {
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
}

func (this *V1Endpoints) GetFilteredHistoricProcessInstances(config configuration.Config, router *http.ServeMux, c *camunda.Camunda, e *events.Events, m Metrics) {
	router.HandleFunc("GET /history/filtered/process-instance", func(writer http.ResponseWriter, request *http.Request) {
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
}

func (this *V1Endpoints) GetFinishedProcessInstances(config configuration.Config, router *http.ServeMux, c *camunda.Camunda, e *events.Events, m Metrics) {
	router.HandleFunc("GET /history/finished/process-instance", func(writer http.ResponseWriter, request *http.Request) {
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
}

func (this *V1Endpoints) FindFinishedProcessInstances(config configuration.Config, router *http.ServeMux, c *camunda.Camunda, e *events.Events, m Metrics) {
	router.HandleFunc("GET /history/finished/process-instance/{searchtype}/{searchvalue}/{limit}/{offset}/{sortby}/{sortdirection}", func(writer http.ResponseWriter, request *http.Request) {
		searchvalue := request.PathValue("searchvalue")
		searchtype := request.PathValue("searchtype")
		limit := request.PathValue("limit")
		offset := request.PathValue("offset")
		sortby := request.PathValue("sortby")
		sortdirection := request.PathValue("sortdirection")

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
}

func (this *V1Endpoints) FindUnfinishedProcessInstances(config configuration.Config, router *http.ServeMux, c *camunda.Camunda, e *events.Events, m Metrics) {
	router.HandleFunc("GET /history/unfinished/process-instance/{searchtype}/{searchvalue}/{limit}/{offset}/{sortby}/{sortdirection}", func(writer http.ResponseWriter, request *http.Request) {
		searchvalue := request.PathValue("searchvalue")
		searchtype := request.PathValue("searchtype")
		limit := request.PathValue("limit")
		offset := request.PathValue("offset")
		sortby := request.PathValue("sortby")
		sortdirection := request.PathValue("sortdirection")

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
}

func (this *V1Endpoints) GetUnfinishedProcessInstances(config configuration.Config, router *http.ServeMux, c *camunda.Camunda, e *events.Events, m Metrics) {
	router.HandleFunc("GET /history/unfinished/process-instance", func(writer http.ResponseWriter, request *http.Request) {
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
}

func (this *V1Endpoints) GetHistoricProcessInstancesOfProcessDefinition(config configuration.Config, router *http.ServeMux, c *camunda.Camunda, e *events.Events, m Metrics) {
	router.HandleFunc("GET /history/process-definition/{id}/process-instance", func(writer http.ResponseWriter, request *http.Request) {
		//"/engine-rest/history/process-instance?processDefinitionId="
		id := request.PathValue("id")

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
}

func (this *V1Endpoints) GetFinishedProcessInstancesOfProcessDefinition(config configuration.Config, router *http.ServeMux, c *camunda.Camunda, e *events.Events, m Metrics) {
	router.HandleFunc("GET /history/process-definition/{id}/process-instance/finished", func(writer http.ResponseWriter, request *http.Request) {
		//"/engine-rest/history/process-instance?processDefinitionId="
		id := request.PathValue("id")

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
}

func (this *V1Endpoints) GetUnfinishedProcessInstancesOfProcessDefinition(config configuration.Config, router *http.ServeMux, c *camunda.Camunda, e *events.Events, m Metrics) {
	router.HandleFunc("GET /history/process-definition/{id}/process-instance/unfinished", func(writer http.ResponseWriter, request *http.Request) {
		//"/engine-rest/history/process-instance?processDefinitionId="
		id := request.PathValue("id")

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
}

func (this *V1Endpoints) DeleteHistoricPricessInstance(config configuration.Config, router *http.ServeMux, c *camunda.Camunda, e *events.Events, m Metrics) {
	router.HandleFunc("DELETE /history/process-instance/{id}", func(writer http.ResponseWriter, request *http.Request) {
		//DELETE "/engine-rest/history/process-instance/" + processInstanceId
		id := request.PathValue("id")

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
}

func (this *V1Endpoints) DeleteProcessInstance(config configuration.Config, router *http.ServeMux, c *camunda.Camunda, e *events.Events, m Metrics) {
	router.HandleFunc("DELETE /process-instance/{id}", func(writer http.ResponseWriter, request *http.Request) {
		//DELETE "/engine-rest/process-instance/" + processInstanceId
		id := request.PathValue("id")

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
