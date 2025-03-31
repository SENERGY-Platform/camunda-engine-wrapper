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

package cleanup

import (
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/client"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/configuration"
)

func RemoveVid(config configuration.Config, ids []string) error {
	c := client.New("http://localhost:" + config.ServerPort)
	for _, id := range ids {
		err, _ := c.DeleteDeployment(client.InternalAdminToken, "", id)
		if err != nil {
			return err
		}
	}
	return nil
}
