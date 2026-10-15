/*
 * Copyright 2026 InfAI (CC SES)
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

package configuration

import (
	"io"
	"os"
	"strings"
	"testing"
)

func TestHandleEnvironmentVarsLogsNoSecret(t *testing.T) {
	t.Setenv("CAMUNDA_DB_PW", "secret-db-password")
	t.Setenv("AUTH_CLIENT_SECRET", "secret-client-secret")
	t.Setenv("WRAPPER_DB", "postgres://user:secret-wrapper@db/wrapper")
	t.Setenv("SHARDING_DB", "postgres://user:secret-sharding@db/sharding")
	t.Setenv("CAMUNDA_DB", "postgres://user:secret-camunda@db/camunda")
	t.Setenv("CAMUNDA_DB_USER", "secret-db-user")
	t.Setenv("INCIDENT_API_URL", "http://incident-api:8080")

	stdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w
	defer func() { os.Stdout = stdout }()
	config := &Config{}
	handleEnvironmentVars(config)
	os.Stdout = stdout
	w.Close()
	out, err := io.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}

	if strings.Contains(string(out), "secret-") {
		t.Errorf("secret in log: %s", out)
	}
	if !strings.Contains(string(out), "INCIDENT_API_URL") {
		t.Errorf("non-secret value not logged: %s", out)
	}
	if config.CamundaDbPw != "secret-db-password" || config.AuthClientSecret != "secret-client-secret" {
		t.Error("secret values not loaded")
	}
}
