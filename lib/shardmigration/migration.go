package shardmigration

import (
	"encoding/json"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/cache"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/shards"
	"log"
	"net/http"
	"strconv"
)

func Run(camundaUrl string, pgConnStr string, limit int) (err error) {
	s, err := shards.New(pgConnStr, cache.None)
	if err != nil {
		return err
	}
	offset := 0
	count := limit
	tenantSet := map[string]bool{}
	log.Println("load tenants from camunda deployments")
	for count == limit {
		tenants, err := getDeploymentTenants(camundaUrl, limit, offset)
		if err != nil {
			return err
		}
		count = len(tenants)
		offset = offset + limit
		for _, tenant := range tenants {
			tenantSet[tenant] = true
		}
	}

	log.Println("ensure entry of", camundaUrl, " in Shard table")
	err = s.EnsureShard(camundaUrl)
	if err != nil {
		return err
	}

	log.Println("map", len(tenantSet), "tenants to", camundaUrl)
	for tenant, _ := range tenantSet {
		log.Println("add", tenant, "to", camundaUrl)
		err = s.SetShardForUser(tenant, camundaUrl)
		if err != nil {
			return err
		}
	}
	log.Println("done")
	return nil
}

type TenantWrapper struct {
	TenantId string `json:"tenantId"`
}

func getDeploymentTenants(camundaUrl string, limit int, offset int) (tenants []string, err error) {
	resp, err := http.Get(camundaUrl + "/engine-rest/deployment?firstResult=" + strconv.Itoa(offset) + "&maxResults=" + strconv.Itoa(limit))
	if err != nil {
		return tenants, err
	}
	defer resp.Body.Close()
	temp := []TenantWrapper{}
	err = json.NewDecoder(resp.Body).Decode(&temp)
	if err != nil {
		return nil, err
	}
	for _, tenant := range temp {
		tenants = append(tenants, tenant.TenantId)
	}
	return tenants, nil
}
