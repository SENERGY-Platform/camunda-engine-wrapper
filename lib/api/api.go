/*
 * Copyright 2019 InfAI (CC SES)
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
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"reflect"
	"runtime/debug"
	"time"

	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/api/util"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/camunda"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/configuration"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/controller"
	"github.com/SENERGY-Platform/service-commons/pkg/accesslog"
)

//go:generate go tool swag init -o ../../docs --parseDependency -d .. -g api/api.go

type EndpointMethod = func(config configuration.Config, router *http.ServeMux, camunda *camunda.Camunda, ctrl *controller.Controller, m Metrics)

var endpoints = []interface{}{} //list of objects with EndpointMethod

func Start(ctx context.Context, config configuration.Config, camunda *camunda.Camunda, ctrl *controller.Controller, m Metrics) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprint(r))
		}
	}()
	router := GetRouter(config, camunda, ctrl, m)

	timeout, err := time.ParseDuration(config.HttpServerTimeout)
	if err != nil {
		config.GetLogger().Warn("invalid http server timeout --> no timeouts", "error", err)
		err = nil
	}

	readtimeout, err := time.ParseDuration(config.HttpServerReadTimeout)
	if err != nil {
		config.GetLogger().Warn("invalid http server read timeout --> no timeouts", "error", err)
		err = nil
	}

	server := &http.Server{Addr: ":" + config.ServerPort, Handler: router, WriteTimeout: timeout, ReadTimeout: readtimeout}
	go func() {
		config.GetLogger().Info("listening", "address", server.Addr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			debug.PrintStack()
			config.GetLogger().Error("FATAL: http server error", "error", err)
			log.Fatal("FATAL:", err)
		}
	}()
	go func() {
		<-ctx.Done()
		config.GetLogger().Info("api shutdown", "result", server.Shutdown(context.Background()))
	}()
	return
}

// GetRouter doc
// @title         Process-Engine-Wrapper
// @version       0.1
// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html
// @BasePath  /
// @securityDefinitions.apikey Bearer
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.
func GetRouter(config configuration.Config, camunda *camunda.Camunda, ctrl *controller.Controller, m Metrics) http.Handler {
	router := http.NewServeMux()
	for _, e := range endpoints {
		for name, call := range getEndpointMethods(e) {
			config.GetLogger().Info("adding endpoint", "name", name)
			call(config, router, camunda, ctrl, m)
		}
	}
	handler := util.NewCors(router)
	handler = accesslog.New(handler, accesslog.Options{TrimFormat: config.AccessLogTrimFormat, TrimAttributes: "body"})
	return handler
}

type Metrics interface {
	NotifyEventTrigger()
}

func getEndpointMethods(e interface{}) map[string]EndpointMethod {
	result := map[string]EndpointMethod{}
	objRef := reflect.ValueOf(e)
	methodCount := objRef.NumMethod()
	for i := 0; i < methodCount; i++ {
		m := objRef.Method(i)
		f, ok := m.Interface().(EndpointMethod)
		if ok {
			name := getTypeName(objRef.Type()) + "::" + objRef.Type().Method(i).Name
			result[name] = f
		}
	}
	return result
}

func getTypeName(t reflect.Type) (res string) {
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t.Name()
}
