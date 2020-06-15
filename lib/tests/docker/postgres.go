package docker

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"log"
	"sync"
)

func Postgres(ctx context.Context, wg *sync.WaitGroup) (conStr string, err error) {
	dbName := "test"
	pool, err := dockertest.NewPool("")
	container, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "postgres",
		Tag:        "11.2",
		Env:        []string{"POSTGRES_DB=" + dbName, "POSTGRES_PASSWORD=pw", "POSTGRES_USER=usr"},
	}, func(config *docker.HostConfig) {
		config.Tmpfs = map[string]string{"/var/lib/postgresql/data": "rw"}
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
	conStr = fmt.Sprintf("postgres://usr:pw@localhost:%s/%s?sslmode=disable", container.GetPort("5432/tcp"), dbName)
	err = pool.Retry(func() error {
		var err error
		log.Println("try connecting to pg")
		db, err := sql.Open("postgres", conStr)
		if err != nil {
			log.Println(err)
			return err
		}
		err = db.Ping()
		if err != nil {
			log.Println(err)
			return err
		}
		return nil
	})
	return
}
