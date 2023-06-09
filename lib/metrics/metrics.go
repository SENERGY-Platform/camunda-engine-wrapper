/*
 * Copyright (c) 2023 InfAI (CC SES)
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

package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"net/http"
)

type Metrics struct {
	EventMessages prometheus.Counter
	httphandler   http.Handler
}

func NewMetrics(prefix string) *Metrics {
	reg := prometheus.NewRegistry()
	m := &Metrics{
		httphandler: promhttp.HandlerFor(
			reg,
			promhttp.HandlerOpts{
				Registry: reg,
			},
		),
		EventMessages: prometheus.NewCounter(prometheus.CounterOpts{
			Name: prefix + "_event_messages",
			Help: "count of event messages received since startup",
		}),
	}

	reg.MustRegister(m.EventMessages)

	return m
}

func (this *Metrics) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	log.Printf("%v [%v] %v \n", request.RemoteAddr, request.Method, request.URL)
	this.httphandler.ServeHTTP(writer, request)
}
