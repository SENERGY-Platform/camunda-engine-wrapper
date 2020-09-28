package docker

import (
	"context"
	"errors"
	"fmt"
	"github.com/ory/dockertest/v3"
	"log"
	"net/http"
	"sync"
)

func Camunda(ctx context.Context, wg *sync.WaitGroup, pgIp string, pgPort string) (camundaUrl string, err error) {
	dbName := "camunda"
	pool, err := dockertest.NewPool("")
	container, err := pool.Run("fgseitsrancher.wifa.intern.uni-leipzig.de:5000/process-engine", "dev", []string{
		"DB_PASSWORD=pw",
		"DB_URL=jdbc:postgresql://" + pgIp + ":" + pgPort + "/" + dbName,
		"DB_PORT=" + pgPort,
		"DB_NAME=" + dbName,
		"DB_HOST=" + pgIp,
		"DB_DRIVER=org.postgresql.Driver",
		"DB_USERNAME=usr",
		"DATABASE=postgres",
	})
	if err != nil {
		return "", err
	}
	wg.Add(1)
	go func() {
		<-ctx.Done()
		log.Println("DEBUG: remove container " + container.Container.Name)
		container.Close()
		wg.Done()
	}()
	ip := container.Container.NetworkSettings.IPAddress
	port := "8080"
	camundaUrl = fmt.Sprintf("http://%s:%s", ip, port)
	err = pool.Retry(func() error {
		log.Println("try camunda connection...")
		resp, err := http.Get(camundaUrl + "/engine-rest/metrics")
		if err != nil {
			return err
		}
		if resp.StatusCode != 200 {
			log.Println("unexpectet response code", resp.StatusCode, resp.Status)
			return errors.New("unexpectet response code: " + resp.Status)
		}
		return nil
	})
	return
}
