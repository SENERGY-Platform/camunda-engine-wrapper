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

import "errors"

//should be called periodically or on startup, to remove all process deployments which are not referenced by a virtual id (remove dead deployments)
//and remove all virtual id relations without process deployments
//TODO: lock on database and camunda access while its running
func clearUnlinkedDeployments() error {

}

//saves relation between vid (command.Id) and deploymentId
func saveVidRelation(command DeploymentCommand, deploymentId string) (err error) {

}

//remove relation between vid (command.Id) and deploymentId
func removeVidRelation(command DeploymentCommand, deploymentId string) (commitFn func(), rollback func(), err error) {

}

//returns deploymentId related to vid
func getDeploymentId(vid string) (deploymentId string, exists bool, err error) {

}

//returns vid related to deploymentId
func getVirtualId(deploymentId string) (vid string, exists bool, err error) {

}

/*
//example for setVid in slices
arr := Deployments{} // alias for []Deployment
for i:=0; i<len(arr); i++ {
	setVid(&arr[i])
}
*/

func setVid(element VidUpdateable) (err error) {
	deploymentId := element.GetDeploymentId()
	vid, exists, err := getVirtualId(deploymentId)
	if err != nil {
		return err
	}
	if !exists {
		return errors.New("no vid found")
	}
	element.SetDeploymentId(vid)
	return nil
}

type VidUpdateable interface {
	SetDeploymentId(id string)
	GetDeploymentId() (id string)
}
