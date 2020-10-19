package docker

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/ory/dockertest"
	dockertestv3 "github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"log"
	"net/http"
)

//deprecated
func Helper_getPgDependency(dbName string) (closer func(), hostPort string, ipAddress string, pgStr string, err error) {
	pool, err := dockertestv3.NewPool("")
	if err != nil {
		return func() {}, "", "", "", err
	}
	log.Println("start postgres db")
	pg, err := pool.RunWithOptions(&dockertestv3.RunOptions{
		Repository: "postgres",
		Tag:        "11.2",
		Env:        []string{"POSTGRES_DB=" + dbName, "POSTGRES_PASSWORD=pw", "POSTGRES_USER=usr"},
	}, func(config *docker.HostConfig) {
		config.Tmpfs = map[string]string{"/var/lib/postgresql/data": "rw"}
	})
	if err != nil {
		return func() {}, "", "", "", err
	}
	var db *sql.DB
	hostPort = pg.GetPort("5432/tcp")
	pgStr = fmt.Sprintf("postgres://usr:pw@localhost:%s/%s?sslmode=disable", hostPort, dbName)
	err = pool.Retry(func() error {
		log.Println("try pg connection...")
		var err error
		db, err = sql.Open("postgres", pgStr)
		if err != nil {
			return err
		}
		return db.Ping()
	})
	return func() { pg.Close() }, hostPort, pg.Container.NetworkSettings.IPAddress, pgStr, err
}

//deprecated
func Helper_getCamundaDependency(pgIp string, pgPort string) (closer func(), hostPort string, ipAddress string, err error) {
	pool, err := dockertest.NewPool("")
	if err != nil {
		return func() {}, "", "", err
	}
	log.Println("start process engine")
	camunda, err := pool.Run("fgseitsrancher.wifa.intern.uni-leipzig.de:5000/process-engine", "dev", []string{
		"DB_PASSWORD=pw",
		"DB_URL=jdbc:postgresql://" + pgIp + ":" + pgPort + "/camunda",
		"DB_PORT=" + pgPort,
		"DB_NAME=camunda",
		"DB_HOST=" + pgIp,
		"DB_DRIVER=org.postgresql.Driver",
		"DB_USERNAME=usr",
		"DATABASE=postgres",
	})
	if err != nil {
		return func() {}, "", "", err
	}
	hostPort = camunda.GetPort("8080/tcp")
	err = pool.Retry(func() error {
		log.Println("try camunda connection...")
		resp, err := http.Get("http://localhost:" + hostPort + "/engine-rest/metrics")
		if err != nil {
			return err
		}
		if resp.StatusCode != 200 {
			log.Println("unexpectet response code", resp.StatusCode, resp.Status)
			return errors.New("unexpectet response code: " + resp.Status)
		}
		return nil
	})
	return func() { camunda.Close() }, hostPort, camunda.Container.NetworkSettings.IPAddress, err
}
