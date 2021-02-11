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
)

func FindUnlinkedVid(config configuration.Config) (unlinkedVid []string, err error) {
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

	deplIndex := map[string]bool{}

	for _, shard := range shards {
		deployments, err := getDeploymentListAllRaw(shard)
		if err != nil {
			return nil, err
		}
		for _, depl := range deployments {
			deplIndex[depl.Id] = true
		}
	}

	for did, vid := range byDeplId {
		if filled, exists := deplIndex[did]; !filled || !exists {
			unlinkedVid = append(unlinkedVid, vid)
		}
	}

	return unlinkedVid, nil
}
