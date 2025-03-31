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
	"fmt"
	"io"
	"net/http"
	"net/url"
)

func RemovePid(ids map[string][]string) error {
	for shard, pids := range ids {
		for _, pid := range pids {
			err := removeProcess(shard, pid)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func removeProcess(shard string, deploymentId string) (err error) {
	anonymousShard, _ := url.Parse(shard)
	anonymousShard.User = &url.Userinfo{}
	fmt.Println("remove pid", anonymousShard.String(), deploymentId)
	url := shard + "/engine-rest/deployment/" + deploymentId + "?cascade=true&skipIoMappings=true"
	request, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	temp, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return errors.New(fmt.Sprint("unable to delete ", anonymousShard.String(), " ", deploymentId, " ", resp.StatusCode, " ", string(temp)))
	}
	return
}
