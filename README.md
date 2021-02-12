## Configuration


| config.json              | env                      | desc                                                                                                                      |
|--------------------------|--------------------------|---------------------------------------------------------------------------------------------------------------------------|
| server_port                | SERVER_PORT               | port of wrapper api                                                                                             |
| zookeeper_url              | ZOOKEEPER_URL             | url to zookeeper                                                                                                          |
| kafka_group                | KAFKA_GROUP               | used kafka consumer group                                                                                        |
| deployment_topic           | DEPLOYMENT_TOPIC          | kafka topic of deployments                                                                                        |
| incident_topic             | INCIDENT_TOPIC            | kafka topic of incidents                                                                                                                          |
| wrapper_db                 | WRAPPER_DB                | connection string to postgres database to store virtual ids (e.g. postgres://usr:pw@databasip:5432/shards?sslmode=disable)                                                                                                                         |
| sharding_db                | SHARDING_DB               | connection string to postgres database to store sharding information (e.g. postgres://usr:pw@databasip:5432/shards?sslmode=disable)                                                                                                                         |
| debug                      | DEBUG                     | more logs                                                      |

## Wrapper
to run the api wrapper normally, call the program without any additional flags
```
./camunda-engine-wrapper
```

## New Shard
- ensure that the config-variable `sharding_db` is set (env or json)
- call `./addshard http://shard-url:8080`

## Vid Consistence Cleanup
- use the cleanup executable to find and remove unlinked vid and processes
- sub commands are
    - list-unlinked-vid: lists unlinked vid 
    - list-unlinked-pid: lists unlinked processes
    - remove-vid: removes given vid
    - remove-pid: removes given processes
- all config variables (except `ServerPort` and `LogLevel` ar used)

```
./cleanup remove-vid $(./cleanup list-unlinked-vid)
```

## Docker
- env variables can be passed with the `-e` flag
- program-flags can be passed at the end of the docker run command 

```
docker build -t enginewrapper .
docker run -e PG_CONN=postgres://usr:pw@databasip:5432/shards?sslmode=disable enginewrapper -migrate_shard=http://new_shard_url:8080
```