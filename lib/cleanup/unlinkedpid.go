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
	"errors"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/configuration"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/shards"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/shards/cache"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/vid"
	"time"
)

func FindUnlinkedPid(config configuration.Config, deploymentAgeBuffer time.Duration) (unlinkedPid map[string][]string, err error) {
	/*
		//test output
		return map[string][]string{
			"test-shard-1": []string{
				"pid1", "pid2", "pid3",
			},
			"test-shard-3": []string{
				"pid3", "pid4", "pid5", "pid6",
			},
		}, nil
	*/
	unlinkedPid = map[string][]string{}
	s, err := shards.New(config.ShardingDb, cache.None)
	if err != nil {
		return nil, err
	}

	v, err := vid.New(config.WrapperDb)
	if err != nil {
		return nil, err
	}

	shards, err := s.GetShards()
	if err != nil {
		return nil, err
	}

	_, byDeplId, err := v.GetRelations()
	if err != nil {
		return nil, err
	}

	if len(shards) == 0 {
		return nil, errors.New("no shards found")
	}

	for _, shard := range shards {
		deployments, err := getDeploymentListAllRaw(shard)
		if err != nil {
			return nil, err
		}
		for _, depl := range deployments {
			camundaTimeFormat := "2006-01-02T15:04:05.000Z0700"
			deplTime, err := time.Parse(camundaTimeFormat, depl.DeploymentTime)
			if err != nil {
				return nil, err
			}
			age := time.Since(deplTime)
			if _, ok := byDeplId[depl.Id]; !ok && age > deploymentAgeBuffer {
				unlinkedPid[shard] = append(unlinkedPid[shard], depl.Id)
			}
		}
	}

	return unlinkedPid, nil
}
