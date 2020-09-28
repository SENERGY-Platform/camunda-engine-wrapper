package vidcleanup

import (
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/cache"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/kafka"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/shards"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/vid"
	"log"
	"time"
)

//should be called periodically or on startup, to remove all process deployments which are not referenced by a virtual id (remove dead deployments)
//and remove all virtual id relations without process deployments
//can lead to a lost update, if called while a process is deploying using a other instance of this program
//to ensure that no deployments are deleted which are in the process of being created, the deploymentAgeBuffer can be used to define a minimal age before a deployment may be deleted
func ClearUnlinkedDeployments(pgConn string, deploymentTopic string, cqrs kafka.Interface, deploymentAgeBuffer time.Duration) error {
	log.Println("clear unlinked deployments")
	s, err := shards.New(pgConn, cache.None)
	if err != nil {
		return err
	}

	v, err := vid.New(pgConn)
	if err != nil {
		return err
	}

	shards, err := s.GetShards()
	if err != nil {
		return err
	}

	_, byDeplId, err := v.GetRelations()
	if err != nil {
		return err
	}
	deplIndex := map[string]bool{}

	for _, shard := range shards {
		log.Println("check shard", shard)
		deployments, err := getDeploymentListAllRaw(shard)
		if err != nil {
			return err
		}
		for _, depl := range deployments {
			camundaTimeFormat := "2006-01-02T15:04:05.000Z0700"
			deplTime, err := time.Parse(camundaTimeFormat, depl.DeploymentTime)
			if err != nil {
				return err
			}
			age := time.Since(deplTime)
			if _, ok := byDeplId[depl.Id]; !ok && age > deploymentAgeBuffer {
				err = removeProcess(depl.Id, shard)
				if err != nil {
					log.Println("WARNING: unable to remove process while clearing unlinked deployments: ", depl.Id, depl.Name, err)
				}
			}
			deplIndex[depl.Id] = true
		}
	}
	for did, vid := range byDeplId {
		if filled, exists := deplIndex[did]; !filled || !exists {
			err = removeVidByEvent(cqrs, deploymentTopic, vid)
			if err != nil {
				log.Println("WARNING: unable to remove process relation while clearing unlinked deployments: ", did, vid, err)
			}
		}
	}
	return nil
}
