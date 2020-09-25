package lib

import (
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/vid"
	"sync"
)

var vidInstance *vid.Vid
var vidOnce sync.Once

func GetVid() (instance *vid.Vid, err error) {
	vidOnce.Do(func() {
		vidInstance, err = vid.New(Config.PgConn)
		if err != nil {
			return
		}
	})
	return vidInstance, err
}

func getDeploymentId(vid string) (deploymentId string, exists bool, err error) {
	instance, err := GetVid()
	if err != nil {
		return deploymentId, exists, err
	}
	return instance.GetDeploymentId(vid)
}

func setVid(element vid.VidUpdateable) (err error) {
	instance, err := GetVid()
	if err != nil {
		return err
	}
	return instance.SetVid(element)
}

func removeVidRelation(vid string, deploymentId string) (commit func() error, rollback func() error, err error) {
	instance, err := GetVid()
	if err != nil {
		return commit, rollback, err
	}
	return instance.RemoveVidRelation(vid, deploymentId)
}

func saveVidRelation(vid string, deploymentId string) (err error) {
	instance, err := GetVid()
	if err != nil {
		return err
	}
	return instance.SaveVidRelation(vid, deploymentId)
}

func vidExists(vid string) (exists bool, err error) {
	instance, err := GetVid()
	if err != nil {
		return exists, err
	}
	return instance.VidExists(vid)
}
