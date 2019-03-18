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

package lib

import (
	"database/sql"
	"errors"
	"log"
	"sync"
)

var maintenanceLock = sync.RWMutex{}

//should be called periodically or on startup, to remove all process deployments which are not referenced by a virtual id (remove dead deployments)
//and remove all virtual id relations without process deployments
//can lead to a lost update, if called while a process is deploying using a other instance of this program
func ClearUnlinkedDeployments() error {
	maintenanceLock.Lock()
	defer maintenanceLock.Unlock()
	db, err := GetDB()
	if err != nil {
		return err
	}
	_, byDeplId, err := getVidRelations(db)
	if err != nil {
		return err
	}
	deployments, err := getDeploymentListAllRaw()
	if err != nil {
		return err
	}
	deplIndex := map[string]bool{}

	for _, depl := range deployments {
		if _, ok := byDeplId[depl.Id]; !ok {
			err = RemoveProcess(depl.Id)
			if err != nil {
				log.Println("WARNING: unable to remove process while clearing unlinked deployments: ", depl.Id, depl.Name, err)
			}
		}
		deplIndex[depl.Id] = true
	}

	for did, vid := range byDeplId {
		if filled, exists := deplIndex[did]; !filled || !exists {
			err = removeVidByEvent(vid)
			if err != nil {
				log.Println("WARNING: unable to remove process relation while clearing unlinked deployments: ", did, vid, err)
			}
		}
	}
	return nil
}

func removeVidByEvent(vid string) error {
	return PublishDeploymentDelete(vid)
}

func getVidRelations(db DbInterface) (byVid map[string]string, byDeploymentId map[string]string, err error) {
	byVid = map[string]string{}
	byDeploymentId = map[string]string{}
	query := `SELECT DeploymentId, VirtualId FROM VidRelation;`
	rows, err := db.Query(query)
	defer rows.Close()
	for rows.Next() {
		var vid string
		var deplId string
		err = rows.Scan(&deplId, &vid)
		if err != nil {
			return
		}
		byVid[vid] = deplId
		byDeploymentId[deplId] = vid
	}
	return
}

//saves relation between vid (command.Id) and deploymentId
func saveVidRelation(vid string, deploymentId string) (err error) {
	db, err := GetDB()
	if err != nil {
		log.Println("ERROR: ", err)
		return err
	}
	_, err = db.Exec("INSERT INTO VidRelation (DeploymentId, VirtualId) VALUES ($1, $2);", deploymentId, vid)
	return err
}

func vidExists(vid string) (exists bool, err error) {
	db, err := GetDB()
	if err != nil {
		log.Println("ERROR: ", err)
		return false, err
	}
	row := db.QueryRow("SELECT COUNT(1) FROM VidRelation WHERE VirtualId = $1;", vid)
	count := 0
	err = row.Scan(&count)
	if err != nil {
		return false, err
	}
	if count > 0 {
		return true, nil
	}
	return false, nil
}

//remove relation between vid (command.Id) and deploymentId
func removeVidRelation(vid string, deploymentId string) (commit func() error, rollback func() error, err error) {
	db, err := GetDB()
	if err != nil {
		return commit, rollback, err
	}
	tx, err := db.Begin()
	if err != nil {
		return commit, rollback, err
	}
	_, err = tx.Exec("DELETE FROM VidRelation WHERE DeploymentId = $1;", deploymentId)
	if err != nil {
		tx.Rollback()
		return commit, rollback, err
	}
	_, err = tx.Exec("DELETE FROM VidRelation WHERE VirtualId = $1; ", vid)
	if err != nil {
		tx.Rollback()
		return commit, rollback, err
	}
	return tx.Commit, tx.Rollback, err
}

//returns deploymentId related to vid
func getDeploymentId(vid string) (deploymentId string, exists bool, err error) {
	exists = false
	db, err := GetDB()
	if err != nil {
		return deploymentId, exists, err
	}
	query := `SELECT DeploymentId FROM VidRelation WHERE VirtualId = $1;`
	rows, err := db.Query(query, vid)
	if err != nil {
		return deploymentId, exists, err
	}
	arr, err := rowsToStringList(rows)
	if len(arr) >= 1 {
		exists = true
	} else {
		exists = false
		return
	}
	return arr[0], exists, err
}

//expects rows with a single value
func rowsToStringList(rows *sql.Rows) (result []string, err error) {
	defer rows.Close()
	for rows.Next() {
		var value string
		err = rows.Scan(&value)
		if err != nil {
			return result, err
		}
		result = append(result, value)
	}
	return
}

//returns vid related to deploymentId
func getVirtualId(deploymentId string) (vid string, exists bool, err error) {
	exists = false
	db, err := GetDB()
	if err != nil {
		return vid, exists, err
	}
	query := `SELECT VirtualId FROM VidRelation WHERE DeploymentId = $1;`
	rows, err := db.Query(query, deploymentId)
	if err != nil {
		return vid, exists, err
	}
	arr, err := rowsToStringList(rows)
	if len(arr) >= 1 {
		exists = true
	} else {
		exists = false
		return
	}
	return arr[0], exists, err
}

/*
//example for setVid in slices
arr := Deployments{} // alias for []Deployment
for i:=0; i<len(arr); i++ {
	setVid(&arr[i])
}
*/

//replaces deployment ids in element with vid from database
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
