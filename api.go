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

package main

import (
	"log"
	"net/http"

	"io"

	"github.com/SmartEnergyPlatform/jwt-http-router"
	"github.com/SmartEnergyPlatform/util/http/cors"
	"github.com/SmartEnergyPlatform/util/http/logger"
	"github.com/SmartEnergyPlatform/util/http/response"
)

func InitApi() {
	log.Println("start server on port: ", Config.ServerPort)
	httpHandler := getRoutes()
	corsHandler := cors.New(httpHandler)
	logger := logger.New(corsHandler, Config.LogLevel)
	log.Println(http.ListenAndServe(":"+Config.ServerPort, logger))
}

func getRoutes() *jwt_http_router.Router {
	router := jwt_http_router.New(jwt_http_router.JwtConfig{PubRsa: Config.JwtPubRsa, ForceAuth: Config.ForceAuth == "true", ForceUser: Config.ForceUser == "true"})

	router.GET("/", func(writer http.ResponseWriter, request *http.Request, params jwt_http_router.Params, jwt jwt_http_router.Jwt) {
		response.To(writer).Json(map[string]string{"status": "OK"})
	})

	router.GET("/process-definition/:id/start", func(writer http.ResponseWriter, request *http.Request, params jwt_http_router.Params, jwt jwt_http_router.Jwt) {
		id := params.ByName("id")
		if err := checkProcessDefinitionAccess(id, jwt.UserId); err != nil {
			log.Println("WARNING: Access denied for user;", jwt.UserId, err)
			http.Error(writer, "Access denied", http.StatusUnauthorized)
			return
		}
		result, err := startProcess(id)
		if err != nil {
			log.Println("ERROR: error on process start", err)
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		response.To(writer).Json(result)
	})

	router.GET("/process-definition/:id/start/id", func(writer http.ResponseWriter, request *http.Request, params jwt_http_router.Params, jwt jwt_http_router.Jwt) {
		id := params.ByName("id")
		if err := checkProcessDefinitionAccess(id, jwt.UserId); err != nil {
			log.Println("WARNING: Access denied for user;", jwt.UserId, err)
			http.Error(writer, "Access denied", http.StatusUnauthorized)
			return
		}
		result, err := startProcessGetId(id)
		if err != nil {
			log.Println("ERROR: error on process start", err)
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		response.To(writer).Json(result)
	})

	router.GET("/deployment/:id", func(writer http.ResponseWriter, request *http.Request, params jwt_http_router.Params, jwt jwt_http_router.Jwt) {
		//"/engine-rest/deployment/" + id
		id := params.ByName("id")
		if err := checkDeploymentAccess(id, jwt.UserId); err != nil {
			log.Println("WARNING: Access denied for user;", jwt.UserId, err)
			http.Error(writer, "Access denied", http.StatusUnauthorized)
			return
		}
		result, err := getDeployment(id)
		if err != nil {
			log.Println("ERROR: error on getDeployment", err)
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		response.To(writer).Json(result)
	})

	router.GET("/deployment/:id/definition", func(writer http.ResponseWriter, request *http.Request, params jwt_http_router.Params, jwt jwt_http_router.Jwt) {
		//"/engine-rest/process-definition?deploymentId=
		id := params.ByName("id")
		if err := checkDeploymentAccess(id, jwt.UserId); err != nil {
			log.Println("WARNING: Access denied for user;", jwt.UserId, id, err)
			http.Error(writer, "Access denied", http.StatusUnauthorized)
			return
		}
		result, err := getDefinitionByDeployment(id)
		if err != nil {
			log.Println("ERROR: error on getDeploymentByDef", err)
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		response.To(writer).Json(result)
	})

	router.GET("/deployment", func(writer http.ResponseWriter, request *http.Request, params jwt_http_router.Params, jwt jwt_http_router.Jwt) {
		// "/engine-rest/deployment?tenantIdIn="+userId
		result, err := getDeploymentList(jwt.UserId, request.URL.Query())
		if err != nil {
			log.Println("ERROR: error on getDeploymentList", err)
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		response.To(writer).Json(result)
	})

	router.GET("/process-definition/:id", func(writer http.ResponseWriter, request *http.Request, params jwt_http_router.Params, jwt jwt_http_router.Jwt) {
		//"/engine-rest/process-definition/" + processDefinitionId
		id := params.ByName("id")
		if err := checkProcessDefinitionAccess(id, jwt.UserId); err != nil {
			log.Println("WARNING: Access denied for user;", jwt.UserId, err)
			http.Error(writer, "Access denied", http.StatusUnauthorized)
			return
		}
		result, err := getProcessDefinition(id)
		if err != nil {
			log.Println("ERROR: error on getProcessDefinition", err)
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		response.To(writer).Json(result)
	})

	router.GET("/process-definition/:id/diagram", func(writer http.ResponseWriter, request *http.Request, params jwt_http_router.Params, jwt jwt_http_router.Jwt) {
		// "/engine-rest/process-definition/" + processDefinitionId + "/diagram"
		id := params.ByName("id")
		if err := checkProcessDefinitionAccess(id, jwt.UserId); err != nil {
			log.Println("WARNING: Access denied for user;", jwt.UserId, err)
			http.Error(writer, "Access denied", http.StatusUnauthorized)
			return
		}
		result, err := getProcessDefinitionDiagram(id)
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

	router.GET("/process-definition/:id/incident/count", func(writer http.ResponseWriter, request *http.Request, params jwt_http_router.Params, jwt jwt_http_router.Jwt) {
		//"/engine-rest/incident/count?processDefinitionId=" + deployment.processDefintionId
		id := params.ByName("id")
		if err := checkProcessDefinitionAccess(id, jwt.UserId); err != nil {
			log.Println("WARNING: Access denied for user;", jwt.UserId, err)
			http.Error(writer, "Access denied", http.StatusUnauthorized)
			return
		}
		result, err := getProcessDefinitionIncident(id)
		if err != nil {
			log.Println("ERROR: error on getProcessDefinitionIncident", err)
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		response.To(writer).Json(result)
	})

	router.GET("/process-instance", func(writer http.ResponseWriter, request *http.Request, params jwt_http_router.Params, jwt jwt_http_router.Jwt) {
		//"/engine-rest/process-instance"
		result, err := getProcessInstanceList(jwt.UserId)
		if err != nil {
			log.Println("ERROR: error on getProcessInstanceList", err)
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		response.To(writer).Json(result)
	})

	router.GET("/process-instances/count", func(writer http.ResponseWriter, request *http.Request, params jwt_http_router.Params, jwt jwt_http_router.Jwt) {
		//"/engine-rest/process-instance/count"
		result, err := getProcessInstanceCount(jwt.UserId)
		if err != nil {
			log.Println("ERROR: error on getProcessInstanceCount", err)
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		response.To(writer).Json(result)
	})

	router.GET("/history/process-instance", func(writer http.ResponseWriter, request *http.Request, params jwt_http_router.Params, jwt jwt_http_router.Jwt) {
		//"/engine-rest/history/process-instance"
		result, err := getProcessInstanceHistoryList(jwt.UserId)
		if err != nil {
			log.Println("ERROR: error on getProcessDefinitionIncident", err)
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		response.To(writer).Json(result)
	})

	router.GET("/history/finished/process-instance", func(writer http.ResponseWriter, request *http.Request, params jwt_http_router.Params, jwt jwt_http_router.Jwt) {
		//"/engine-rest/history/process-instance"
		result, err := getProcessInstanceHistoryListFinished(jwt.UserId)
		if err != nil {
			log.Println("ERROR: error on getProcessDefinitionIncident", err)
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		response.To(writer).Json(result)
	})

	router.GET("/history/unfinished/process-instance", func(writer http.ResponseWriter, request *http.Request, params jwt_http_router.Params, jwt jwt_http_router.Jwt) {
		//"/engine-rest/history/process-instance"
		result, err := getProcessInstanceHistoryListUnfinished(jwt.UserId)
		if err != nil {
			log.Println("ERROR: error on getProcessDefinitionIncident", err)
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		response.To(writer).Json(result)
	})

	router.GET("/history/process-definition/:id/process-instance", func(writer http.ResponseWriter, request *http.Request, params jwt_http_router.Params, jwt jwt_http_router.Jwt) {
		//"/engine-rest/history/process-instance?processDefinitionId="
		id := params.ByName("id")
		if err := checkProcessDefinitionAccess(id, jwt.UserId); err != nil {
			log.Println("WARNING: Access denied for user;", jwt.UserId, err)
			http.Error(writer, "Access denied", http.StatusUnauthorized)
			return
		}
		result, err := getProcessInstanceHistoryByProcessDefinition(id)
		if err != nil {
			log.Println("ERROR: error on processinstanceHistoryByDefinition", err)
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		response.To(writer).Json(result)
	})

	router.GET("/history/process-definition/:id/process-instance/finished", func(writer http.ResponseWriter, request *http.Request, params jwt_http_router.Params, jwt jwt_http_router.Jwt) {
		log.Println("debug: list finished")
		//"/engine-rest/history/process-instance?processDefinitionId="
		id := params.ByName("id")
		if err := checkProcessDefinitionAccess(id, jwt.UserId); err != nil {
			log.Println("WARNING: Access denied for user;", jwt.UserId, err)
			http.Error(writer, "Access denied", http.StatusUnauthorized)
			return
		}
		result, err := getProcessInstanceHistoryByProcessDefinitionFinished(id)
		if err != nil {
			log.Println("ERROR: error on processinstanceHistoryByDefinition", err)
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		response.To(writer).Json(result)
	})

	router.GET("/history/process-definition/:id/process-instance/unfinished", func(writer http.ResponseWriter, request *http.Request, params jwt_http_router.Params, jwt jwt_http_router.Jwt) {
		//"/engine-rest/history/process-instance?processDefinitionId="
		id := params.ByName("id")
		if err := checkProcessDefinitionAccess(id, jwt.UserId); err != nil {
			log.Println("WARNING: Access denied for user;", jwt.UserId, err)
			http.Error(writer, "Access denied", http.StatusUnauthorized)
			return
		}
		result, err := getProcessInstanceHistoryByProcessDefinitionUnfinished(id)
		if err != nil {
			log.Println("ERROR: error on processinstanceHistoryByDefinition", err)
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		response.To(writer).Json(result)
	})

	router.GET("/history/process-instance/:id/variable-instance", func(writer http.ResponseWriter, request *http.Request, params jwt_http_router.Params, jwt jwt_http_router.Jwt) {
		//"/engine-rest/history/variable-instance?processInstanceId=" + processInstanceId
		id := params.ByName("id")
		if err := checkHistoryAccess(id, jwt.UserId); err != nil {
			log.Println("WARNING: Access denied for user;", jwt.UserId, err)
			http.Error(writer, "Access denied", http.StatusUnauthorized)
			return
		}
		result, err := getProcessInstanceHistoricVariables(id)
		if err != nil {
			log.Println("ERROR: error on getProcessInstanceHistoricVariables", err)
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		response.To(writer).Json(result)
	})

	router.DELETE("/history/process-instance/:id", func(writer http.ResponseWriter, request *http.Request, params jwt_http_router.Params, jwt jwt_http_router.Jwt) {
		//DELETE "/engine-rest/history/process-instance/" + processInstanceId
		id := params.ByName("id")
		if err := checkHistoryAccess(id, jwt.UserId); err != nil {
			log.Println("WARNING: Access denied for user;", jwt.UserId, err)
			http.Error(writer, "Access denied", http.StatusUnauthorized)
			return
		}
		result, err := removeProcessInstanceHistory(id)
		if err != nil {
			log.Println("ERROR: error on removeProcessInstanceHistory", err)
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		response.To(writer).Json(result)
	})

	router.DELETE("/process-instance/:id", func(writer http.ResponseWriter, request *http.Request, params jwt_http_router.Params, jwt jwt_http_router.Jwt) {
		//DELETE "/engine-rest/process-instance/" + processInstanceId
		id := params.ByName("id")
		if err := checkProcessInstanceAccess(id, jwt.UserId); err != nil {
			log.Println("WARNING: Access denied for user;", jwt.UserId, err)
			http.Error(writer, "Access denied", http.StatusUnauthorized)
			return
		}
		err := removeProcessInstance(id)
		if err != nil {
			log.Println("ERROR: error on removeProcessInstance", err)
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		response.To(writer).Text("ok")
	})

	router.GET("/process-instance/:id/incident", func(writer http.ResponseWriter, request *http.Request, params jwt_http_router.Params, jwt jwt_http_router.Jwt) {
		//"/engine-rest/incident?processInstanceId=" + processInstanceId
		id := params.ByName("id")
		if err := checkProcessInstanceAccess(id, jwt.UserId); err != nil {
			log.Println("WARNING: Access denied for user;", jwt.UserId, err)
			http.Error(writer, "Access denied", http.StatusUnauthorized)
			return
		}
		result, err := getProcessInstanceIncidents(id)
		if err != nil {
			log.Println("ERROR: error on getProcessInstanceIncidents", err)
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		response.To(writer).Json(result)
	})

	return router
}
