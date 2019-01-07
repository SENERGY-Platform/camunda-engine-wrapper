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
	"errors"
	"github.com/SmartEnergyPlatform/jwt-http-router"
	"log"
)

func Migrate() error {
	db, err := GetDB()
	if err != nil {
		return err
	}
	relationsByVid, _, err := getVidRelations(db)
	if err != nil {
		return err
	}

	deplExistsExternIndex, err := migrateGetExternDeploymentsIndex()
	if err != nil {
		return err
	}

	deployments, err := getDeploymentListAllRaw()
	if err != nil {
		return err
	}
	deploymentsExistsIntern := map[string]bool{}
	for _, depl := range deployments {
		deploymentsExistsIntern[depl.Id] = true
	}

	for id, metadata := range deplExistsExternIndex {
		relationsDeplId, relationExistsAsVid := relationsByVid[id]
		if relationExistsAsVid {
			exists := deploymentsExistsIntern[relationsDeplId]
			if exists {
				//consistent case => do nothing
			} else {
				//missing deployment => redeploy
				log.Println("DEBUG: ", id, " handleDeploymentCreate()")
				err = handleDeploymentCreate(DeploymentCommand{
					Id:            id,
					Command:       "PUT",
					Owner:         metadata.Owner,
					DeploymentXml: metadata.Abstract.Xml,
					Deployment: DeploymentRequest{
						Svg:     "",
						Process: metadata.Abstract,
					},
				})
			}
		} else {
			exists := deploymentsExistsIntern[id]
			if exists {
				//missing relation (pre vid data) => create relation with vid == deplId
				log.Println("DEBUG: ", id, " saveVidRelation(id, id)")
				err = saveVidRelation(id, id)
			} else {
				//missing deployment and relation => redeploy
				log.Println("DEBUG: ", id, " handleDeploymentCreate()")
				err = handleDeploymentCreate(DeploymentCommand{
					Id:            id,
					Command:       "PUT",
					Owner:         metadata.Owner,
					DeploymentXml: metadata.Abstract.Xml,
					Deployment: DeploymentRequest{
						Svg:     "",
						Process: metadata.Abstract,
					},
				})
			}
		}
		if err != nil {
			return err
		}
	}

	//sync relations and deployments
	err = clearUnlinkedDeployments()
	if err != nil {
		return err
	}

	//delete deployments which dont exist in external service
	relationsByVid, _, err = getVidRelations(db)
	if err != nil {
		return err
	}
	for vid, _ := range relationsByVid {
		_, ok := deplExistsExternIndex[vid]
		if !ok {
			//deployment does not exist in external service => delete
			log.Println("DEBUG: ", vid, " handleDeploymentDelete()")
			err = handleDeploymentDelete(vid)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func migrateGetExternDeploymentsIndex() (result map[string]DeploymentMetadata, err error) {
	if Config.MigrateProcessDeploymentUrl == "" {
		return result, errors.New("missing MigrateProcessDeploymentUrl")
	}
	result = map[string]DeploymentMetadata{}
	deployments := []DeploymentMetadata{}
	err = jwt_http_router.JwtImpersonate(Config.MigrateJwt).GetJSON(Config.MigrateProcessDeploymentUrl+"/deployment", &deployments)
	if err != nil {
		return
	}
	for _, depl := range deployments {
		result[depl.Process] = depl
	}
	return
}
