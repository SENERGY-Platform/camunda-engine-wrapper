package docker

import (
	"context"
	"errors"
	"fmt"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

func Camunda(ctx context.Context, wg *sync.WaitGroup, pgIp string, pgPort string) (camundaUrl string, err error) {
	log.Println("start camunda")
	dbName := "camunda"
	c, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "ghcr.io/senergy-platform/process-engine:dev",
			ExposedPorts: []string{"8080/tcp"},
			WaitingFor: wait.ForAll(
				wait.ForListeningPort("8080/tcp"),
				wait.ForLog("Server initialization in"),
				wait.ForLog("Server startup in"),
			),
			Env: map[string]string{
				"DB_PASSWORD": "pw",
				"DB_URL":      "jdbc:postgresql://" + pgIp + ":" + pgPort + "/" + dbName,
				"DB_PORT":     pgPort,
				"DB_NAME":     dbName,
				"DB_HOST":     pgIp,
				"DB_DRIVER":   "org.postgresql.Driver",
				"DB_USERNAME": "usr",
				"DATABASE":    "postgres",
			},
		},
		Started: true,
	})
	if err != nil {
		return "", err
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		defer func() {
			log.Println("DEBUG: remove container camunda", c.Terminate(context.Background()))
		}()
		<-ctx.Done()
		reader, err := c.Logs(context.Background())
		if err != nil {
			log.Println("ERROR: unable to get container log")
			return
		}
		buf := new(strings.Builder)
		io.Copy(buf, reader)
		fmt.Println("CAMUNDA LOGS: ------------------------------------------")
		fmt.Println(buf.String())
		fmt.Println("\n---------------------------------------------------------------")
	}()

	containerip, err := c.ContainerIP(ctx)
	if err != nil {
		return "", err
	}

	camundaUrl = fmt.Sprintf("http://%s:%s", containerip, "8080")

	err = Retry(time.Minute, func() error {
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

	return camundaUrl, err
}
