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
	"errors"
	"fmt"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/auth"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/camunda"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/configuration"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/controller"
	"log"
	"net/http"
	"time"

	"io"
)

func init() {
	endpoints = append(endpoints, &V2Endpoints{})
}

type V2Endpoints struct{}

// StartProcessDefinition godoc
// @Summary      start process-definitions
// @Description  start process-definitions by id
// @Tags         start, process-definitions
// @Produce      json
// @Security Bearer
// @Param        id path string true "process-definitions id"
// @Success      200
// @Failure      400
// @Failure      401
// @Failure      403
// @Failure      404
// @Failure      500
// @Router       /v2/process-definitions/{id}/start [GET]
func (this *V2Endpoints) StartProcessDefinition(config configuration.Config, router *http.ServeMux, c *camunda.Camunda, e *controller.Controller, m Metrics) {
	router.HandleFunc("GET /v2/process-definitions/{id}/start", func(writer http.ResponseWriter, request *http.Request) {
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

// StartProcessDefinitionAndGetInstanceId godoc
// @Summary      start process-definitions and get instance
// @Description  start process-definitions and get started process-instance
// @Tags         start, process-definitions
// @Produce      json
// @Security Bearer
// @Param        id path string true "process-definitions id"
// @Success      200 {object}  model.ProcessInstance
// @Failure      400
// @Failure      401
// @Failure      403
// @Failure      404
// @Failure      500
// @Router       /v2/process-definitions/{id}/start/id [GET]
func (this *V2Endpoints) StartProcessDefinitionAndGetInstanceId(config configuration.Config, router *http.ServeMux, c *camunda.Camunda, e *controller.Controller, m Metrics) {
	router.HandleFunc("GET /v2/process-definitions/{id}/start/id", func(writer http.ResponseWriter, request *http.Request) {
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

// GetDeployment godoc
// @Summary      get deployment
// @Description  get process deployment by id
// @Tags         deployment
// @Produce      json
// @Security Bearer
// @Param        id path string true "deployment id"
// @Success      200 {object}  model.CamundaDeployment
// @Failure      400
// @Failure      401
// @Failure      403
// @Failure      404
// @Failure      500
// @Router       /v2/deployments/{id} [GET]
func (this *V2Endpoints) GetDeployment(config configuration.Config, router *http.ServeMux, c *camunda.Camunda, e *controller.Controller, m Metrics) {
	router.HandleFunc("GET /v2/deployments/{id}", func(writer http.ResponseWriter, request *http.Request) {
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

// DeploymentExists godoc
// @Summary      deployment exists
// @Description  deployment exists
// @Tags         deployment
// @Produce      json
// @Security Bearer
// @Param        id path string true "deployment id"
// @Success      200 {object}  bool
// @Failure      400
// @Failure      401
// @Failure      403
// @Failure      404
// @Failure      500
// @Router       /v2/deployments/{id}/exists [GET]
func (this *V2Endpoints) DeploymentExists(config configuration.Config, router *http.ServeMux, c *camunda.Camunda, e *controller.Controller, m Metrics) {
	router.HandleFunc("GET /v2/deployments/{id}/exists", func(writer http.ResponseWriter, request *http.Request) {
		id := request.PathValue("id")

		token, err := auth.GetParsedToken(request)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}

		err = c.CheckDeploymentAccess(id, token.GetUserId())
		if errors.Is(err, camunda.UnknownVid) || errors.Is(err, camunda.CamundaDeploymentUnknown) {
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

// StartDeployment godoc
// @Summary      start deployment
// @Description  start deployment by id
// @Tags         start, deployment
// @Produce      json
// @Security Bearer
// @Param        id path string true "deployment id"
// @Success      200 {object}  model.ProcessInstance
// @Failure      400
// @Failure      401
// @Failure      403
// @Failure      404
// @Failure      500
// @Router       /v2/deployments/{id}/start [GET]
func (this *V2Endpoints) StartDeployment(config configuration.Config, router *http.ServeMux, c *camunda.Camunda, e *controller.Controller, m Metrics) {
	router.HandleFunc("GET /v2/deployments/{id}/start", func(writer http.ResponseWriter, request *http.Request) {
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

// GetDeploymentParameters godoc
// @Summary      get deployment parameter
// @Description  get deployment start parameter
// @Tags         start, deployment
// @Produce      json
// @Security Bearer
// @Param        id path string true "deployment id"
// @Success      200 {object}  model.VariableMap
// @Failure      400
// @Failure      401
// @Failure      403
// @Failure      404
// @Failure      500
// @Router       /v2/deployments/{id}/parameter [GET]
func (this *V2Endpoints) GetDeploymentParameters(config configuration.Config, router *http.ServeMux, c *camunda.Camunda, e *controller.Controller, m Metrics) {
	router.HandleFunc("GET /v2/deployments/{id}/parameter", func(writer http.ResponseWriter, request *http.Request) {
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

// GetDeploymentDefinition godoc
// @Summary      get deployment process-definition
// @Description  get deployment process-definition
// @Tags         deployment, process-definition
// @Produce      json
// @Security Bearer
// @Param        id path string true "deployment id"
// @Success      200 {object}  model.ProcessDefinition
// @Failure      400
// @Failure      401
// @Failure      403
// @Failure      404
// @Failure      500
// @Router       /v2/deployments/{id}/definition [GET]
func (this *V2Endpoints) GetDeploymentDefinition(config configuration.Config, router *http.ServeMux, c *camunda.Camunda, e *controller.Controller, m Metrics) {
	router.HandleFunc("GET /v2/deployments/{id}/definition", func(writer http.ResponseWriter, request *http.Request) {
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

// GetDeploymentInstances godoc
// @Summary      get deployment process-instances
// @Description  get deployment process-instances
// @Tags         deployment, process-instance
// @Produce      json
// @Security Bearer
// @Param        id path string true "deployment id"
// @Success      200 {array} model.ProcessInstance
// @Failure      400
// @Failure      401
// @Failure      403
// @Failure      404
// @Failure      500
// @Router       /v2/deployments/{id}/instances [GET]
func (this *V2Endpoints) GetDeploymentInstances(config configuration.Config, router *http.ServeMux, c *camunda.Camunda, e *controller.Controller, m Metrics) {
	router.HandleFunc("GET /v2/deployments/{id}/instances", func(writer http.ResponseWriter, request *http.Request) {
		id := request.PathValue("id")
		token, err := auth.GetParsedToken(request)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		result, err := c.GetInstancesByDeploymentVid(id, token.GetUserId())
		if errors.Is(err, camunda.UnknownVid) {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		writer.Header().Set("Content-Type", "application/json")
		json.NewEncoder(writer).Encode(result)
	})
}

// ListDeployments godoc
// @Summary      list deployments
// @Description  list deployments
// @Tags         deployment
// @Produce      json
// @Security Bearer
// @Success      200 {array}  model.ExtendedDeployment
// @Failure      400
// @Failure      401
// @Failure      403
// @Failure      404
// @Failure      500
// @Router       /v2/deployments [GET]
func (this *V2Endpoints) ListDeployments(config configuration.Config, router *http.ServeMux, c *camunda.Camunda, e *controller.Controller, m Metrics) {
	router.HandleFunc("GET /v2/deployments", func(writer http.ResponseWriter, request *http.Request) {
		token, err := auth.GetParsedToken(request)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		result, err := c.GetExtendedDeploymentList(token.GetUserId(), request.URL.Query())
		if errors.Is(err, camunda.UnknownVid) {
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

// GetProcessDefinition godoc
// @Summary      get process-definition
// @Description  get process-definition
// @Tags         process-definition
// @Produce      json
// @Security Bearer
// @Param        id path string true "process-definition id"
// @Success      200 {object}  model.ProcessDefinition
// @Failure      400
// @Failure      401
// @Failure      403
// @Failure      404
// @Failure      500
// @Router       /v2/process-definitions/{id} [GET]
func (this *V2Endpoints) GetProcessDefinition(config configuration.Config, router *http.ServeMux, c *camunda.Camunda, e *controller.Controller, m Metrics) {
	router.HandleFunc("GET /v2/process-definitions/{id}", func(writer http.ResponseWriter, request *http.Request) {
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

// GetProcessDefinitionDiagram godoc
// @Summary      get process-definition diagram
// @Description  get process-definition diagram
// @Tags         process-definition
// @Produce      json
// @Security Bearer
// @Param        id path string true "process-definition id"
// @Success      200 {object}  string
// @Failure      400
// @Failure      401
// @Failure      403
// @Failure      404
// @Failure      500
// @Router       /v2/process-definitions/{id}/diagram [GET]
func (this *V2Endpoints) GetProcessDefinitionDiagram(config configuration.Config, router *http.ServeMux, c *camunda.Camunda, e *controller.Controller, m Metrics) {
	router.HandleFunc("GET /v2/process-definitions/{id}/diagram", func(writer http.ResponseWriter, request *http.Request) {
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

// ListProcessInstances godoc
// @Summary      list process-instances
// @Description  list process-instances
// @Tags         process-instance
// @Produce      json
// @Security Bearer
// @Success      200 {array}  model.ProcessInstance
// @Failure      400
// @Failure      401
// @Failure      403
// @Failure      404
// @Failure      500
// @Router       /v2/process-instances [GET]
func (this *V2Endpoints) ListProcessInstances(config configuration.Config, router *http.ServeMux, c *camunda.Camunda, e *controller.Controller, m Metrics) {
	router.HandleFunc("GET /v2/process-instances", func(writer http.ResponseWriter, request *http.Request) {
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

// GetProcessInstanceCount godoc
// @Summary      process-instance count
// @Description  process-instance count
// @Tags         process-instance
// @Produce      json
// @Security Bearer
// @Success      200 {object} int
// @Failure      400
// @Failure      401
// @Failure      403
// @Failure      404
// @Failure      500
// @Router       /v2/process-instances/count [GET]
func (this *V2Endpoints) GetProcessInstanceCount(config configuration.Config, router *http.ServeMux, c *camunda.Camunda, e *controller.Controller, m Metrics) {
	router.HandleFunc("GET /v2/process-instances/count", func(writer http.ResponseWriter, request *http.Request) {
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

// GetHistoricProcessInstances godoc
// @Summary      list historic process-instances
// @Description  list historic process-instances
// @Tags         process-instance
// @Produce      json
// @Security Bearer
// @Param        with_total query bool false "if set to true, wraps the result in an objet with the result {total:0, data:[]}"
// @Success      200 {array} model.HistoricProcessInstance
// @Failure      400
// @Failure      401
// @Failure      403
// @Failure      404
// @Failure      500
// @Router       /v2/history/process-instances [GET]
func (this *V2Endpoints) GetHistoricProcessInstances(config configuration.Config, router *http.ServeMux, c *camunda.Camunda, e *controller.Controller, m Metrics) {
	router.HandleFunc("GET /v2/history/process-instances", func(writer http.ResponseWriter, request *http.Request) {
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
			writer.Header().Set("Content-Type", "application/json")
			json.NewEncoder(writer).Encode(result)
		} else {
			result, err := c.GetFilteredProcessInstanceHistoryList(token.GetUserId(), query)
			if err != nil {
				log.Println("ERROR: error on getFilteredProcessInstanceHistoryList", err)
				http.Error(writer, err.Error(), http.StatusInternalServerError)
				return
			}
			writer.Header().Set("Content-Type", "application/json")
			json.NewEncoder(writer).Encode(result)
		}

	})
}

// DeleteHistoricProcessInstance godoc
// @Summary      delete historic process-instance
// @Description  delete historic process-instance
// @Tags         process-instance
// @Security Bearer
// @Param        id path string true "process-instance id"
// @Success      200
// @Failure      400
// @Failure      500
// @Router       /v2/history/process-instances/{id} [DELETE]
func (this *V2Endpoints) DeleteHistoricProcessInstance(config configuration.Config, router *http.ServeMux, c *camunda.Camunda, e *controller.Controller, m Metrics) {
	router.HandleFunc("DELETE /v2/history/process-instances/{id}", func(writer http.ResponseWriter, request *http.Request) {
		id := request.PathValue("id")
		token, err := auth.GetParsedToken(request)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		err, code := e.DeleteHistoricProcessInstance(token.GetUserId(), id)
		if err != nil {
			log.Println("ERROR: error on PublishIncidentDeleteByProcessInstanceEvent", err)
			http.Error(writer, err.Error(), code)
			return
		}
		writer.Header().Set("Content-Type", "application/json")
		json.NewEncoder(writer).Encode("ok")
	})
}

// DeleteProcessInstance godoc
// @Summary      delete process-instance
// @Description  delete process-instance
// @Tags         process-instance
// @Security Bearer
// @Param        id path string true "process-instance id"
// @Success      200
// @Failure      400
// @Failure      500
// @Router       /v2/process-instances/{id} [DELETE]
func (this *V2Endpoints) DeleteProcessInstance(config configuration.Config, router *http.ServeMux, c *camunda.Camunda, e *controller.Controller, m Metrics) {
	router.HandleFunc("DELETE /v2/process-instances/{id}", func(writer http.ResponseWriter, request *http.Request) {
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

// SetProcessInstanceVariable godoc
// @Summary      set process-instance variable
// @Description  set process-instance variable
// @Tags         process-instance
// @Security Bearer
// @Param        id path string true "process-instance id"
// @Param        variable_name path string true "variable_name"
// @Param        message body interface{} true "value"
// @Success      200
// @Failure      400
// @Failure      401
// @Failure      403
// @Failure      404
// @Failure      500
// @Router       /v2/process-instances/{id}/variables/{variable_name} [PUT]
func (this *V2Endpoints) SetProcessInstanceVariable(config configuration.Config, router *http.ServeMux, c *camunda.Camunda, e *controller.Controller, m Metrics) {
	router.HandleFunc("PUT /v2/process-instances/{id}/variables/{variable_name}", func(writer http.ResponseWriter, request *http.Request) {
		id := request.PathValue("id")
		varName := request.PathValue("variable_name")
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
		if err, code := c.CheckProcessInstanceAccess(id, token.GetUserId()); err != nil {
			http.Error(writer, err.Error(), code)
			return
		}
		err = c.SetProcessInstanceVariable(id, token.GetUserId(), varName, varValue)
		if err != nil {
			log.Println("ERROR: error on variable update", err)
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		writer.Header().Set("Content-Type", "application/json")
		json.NewEncoder(writer).Encode("ok")
	})
}

// DeleteProcessMultipleInstances godoc
// @Summary      delete multiple process-instances
// @Description  delete multiple process-instances
// @Tags         process-instance
// @Security Bearer
// @Param        message body []string true "ids"
// @Success      200
// @Failure      400
// @Failure      500
// @Router       /v2/process-instances [DELETE]
func (this *V2Endpoints) DeleteProcessMultipleInstances(config configuration.Config, router *http.ServeMux, c *camunda.Camunda, e *controller.Controller, m Metrics) {
	router.HandleFunc("DELETE /v2/process-instances", func(writer http.ResponseWriter, request *http.Request) {
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
			if err, code := c.CheckProcessInstanceAccess(id, token.GetUserId()); err != nil {
				http.Error(writer, err.Error(), code)
				return
			}
			err := c.RemoveProcessInstance(id, token.GetUserId())
			if err != nil {
				log.Println("ERROR: error on removeProcessInstance", err)
				http.Error(writer, err.Error(), http.StatusInternalServerError)
				return
			}
		}
		writer.Header().Set("Content-Type", "application/json")
		json.NewEncoder(writer).Encode("ok")
		return
	})
}

// DeleteMultipleHistoricProcessInstances godoc
// @Summary      delete multiple historic process-instances
// @Description  delete multiple historic process-instances
// @Tags         process-instance
// @Security Bearer
// @Param        message body []string true "ids"
// @Success      200
// @Failure      400
// @Failure      500
// @Router       /v2/process-instances [DELETE]
func (this *V2Endpoints) DeleteMultipleHistoricProcessInstances(config configuration.Config, router *http.ServeMux, c *camunda.Camunda, e *controller.Controller, m Metrics) {
	router.HandleFunc("DELETE /v2/history/process-instances", func(writer http.ResponseWriter, request *http.Request) {
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
		userId := token.GetUserId()
		for _, id := range ids {
			err, code := e.DeleteHistoricProcessInstance(userId, id)
			if err != nil {
				log.Println("ERROR: error on PublishIncidentDeleteByProcessInstanceEvent", err)
				http.Error(writer, err.Error(), code)
				return
			}
		}
		writer.Header().Set("Content-Type", "application/json")
		json.NewEncoder(writer).Encode("ok")
		return
	})
}

// ref https://docs.camunda.org/rest/camunda-bpm-platform/7.23-SNAPSHOT/#tag/Message/operation/deliverMessage
type EventTriggerMessage struct {
	MessageName           string `json:"messageName"`
	ProcessVariablesLocal map[string]struct {
		Type  string      `json:"type,omitempty"`
		Value interface{} `json:"value,omitempty"`
	} `json:"processVariablesLocal"`
	ResultEnabled bool   `json:"resultEnabled"`
	TenantId      string `json:"tenantId"`
}

// TriggerEvent godoc
// @Summary      trigger event
// @Description  trigger event
// @Tags         event
// @Produce      json
// @Security Bearer
// @Param        message body EventTriggerMessage true "ref https://docs.camunda.org/rest/camunda-bpm-platform/7.23-SNAPSHOT/#tag/Message/operation/deliverMessage"
// @Success      200
// @Failure      400
// @Failure      401
// @Failure      403
// @Failure      404
// @Failure      500
// @Router       /v2/event-trigger [POST]
func (this *V2Endpoints) TriggerEvent(config configuration.Config, router *http.ServeMux, c *camunda.Camunda, e *controller.Controller, m Metrics) {
	router.HandleFunc("POST /v2/event-trigger", func(writer http.ResponseWriter, request *http.Request) {
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
