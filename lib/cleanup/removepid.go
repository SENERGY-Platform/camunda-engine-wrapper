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
	"io/ioutil"
	"net/http"
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
	fmt.Println("remove pid", shard, deploymentId)
	client := &http.Client{}
	url := shard + "/engine-rest/deployment/" + deploymentId + "?cascade=true&skipIoMappings=true"
	request, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}
	resp, err := client.Do(request)
	if err != nil {
		return err
	}
	if resp.StatusCode >= 300 {
		temp, _ := ioutil.ReadAll(resp.Body)
		return errors.New(fmt.Sprint("unable to delete ", shard, " ", deploymentId, " ", resp.StatusCode, " ", string(temp)))
	}
	return
}
