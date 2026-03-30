package shardmigration

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/shards"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/shards/cache"
)

func Add(camundaUrl string, pgConnStr string, batchSize int) (err error) {
	slog.Info("start shard migration", "camundaUrl", camundaUrl, "pgConnStr", pgConnStr, "batchSize", batchSize)
	s, err := shards.New(pgConnStr, cache.None)
	if err != nil {
		return err
	}
	offset := 0
	count := batchSize
	tenantSet := map[string]bool{}
	slog.Debug("load tenants from camunda deployments")
	for count == batchSize {
		tenants, err := getDeploymentTenants(camundaUrl, batchSize, offset)
		if err != nil {
			return err
		}
		count = len(tenants)
		offset = offset + batchSize
		for _, tenant := range tenants {
			tenantSet[tenant] = true
		}
	}

	slog.Debug(fmt.Sprint("ensure entry of", camundaUrl, " in Shard table"))
	err = s.EnsureShard(camundaUrl)
	if err != nil {
		return err
	}

	slog.Debug(fmt.Sprint("map", len(tenantSet), "tenants to", camundaUrl))
	for tenant, _ := range tenantSet {
		slog.Debug(fmt.Sprint("add", tenant, "to", camundaUrl))
		err = s.SetShardForUser(tenant, camundaUrl)
		if err != nil {
			return err
		}
	}
	slog.Debug("done")
	return nil
}

func Remove(camundaUrl string, pgConnStr string) (err error) {
	slog.Info("start shard migration", "camundaUrl", camundaUrl, "pgConnStr", pgConnStr)
	s, err := shards.New(pgConnStr, cache.None)
	if err != nil {
		return err
	}

	slog.Debug(fmt.Sprint("remove entry of", camundaUrl, " in Shard table"))
	err = s.RemoveShard(camundaUrl)
	if err != nil {
		return err
	}

	slog.Debug("done")
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
