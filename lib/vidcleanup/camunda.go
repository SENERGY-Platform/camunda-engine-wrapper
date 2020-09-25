package vidcleanup

import (
	"github.com/SmartEnergyPlatform/util/http/request"
	"net/http"
)

//returns all process deployments without replacing the deployment id with the virtual id
func getDeploymentListAllRaw(shard string) (result Deployments, err error) {
	path := shard + "/engine-rest/deployment"
	err = request.Get(path, &result)
	return
}

func removeProcess(deploymentId string, shard string) (err error) {
	client := &http.Client{}
	url := shard + "/engine-rest/deployment/" + deploymentId + "?cascade=true&skipIoMappings=true"
	request, err := http.NewRequest("DELETE", url, nil)
	_, err = client.Do(request)
	return
}

type Deployment struct {
	Id             string `json:"id"`
	Name           string `json:"name"`
	Source         string `json:"source"`
	DeploymentTime string `json:"deploymentTime"`
	TenantId       string `json:"tenantId"`
}

type Deployments = []Deployment
