## Configuration


| config.json              | env                      | desc                                                                                                                      |
|--------------------------|--------------------------|---------------------------------------------------------------------------------------------------------------------------|
| ServerPort                | SERVER_PORT               | port of wrapper api                                                                                             |
| LogLevel                  | LOG_LEVEL                 | decides which api calls should be logged (DEBUG, CALL, NONE)                                                     |
| ZookeeperUrl              | ZOOKEEPER_URL             | url to zookeeper                                                                                                          |
| KafkaGroup                | KAFKA_GROUP               | used kafka consumer group                                                                                        |
| KafkaDebug                | KAFKA_DEBUG               | more logs from kafka                                                 |
| DeploymentTopic           | DEPLOYMENT_TOPIC          | kafka topic of deployments                                                                                        |
| IncidentTopic             | INCIDENT_TOPIC            | kafka topic of incidents                                                                                                                          |
| PgConn                    | PG_CONN                   | connection string to postgres database                                                                                                                          |
| Debug                     | DEBUG                     | more logs                                                      |

## Wrapper
to run the api wrapper normally, call the program without any additional flags
```
./camunda-engine-wrapper
```

## New Shard
- ensure that the config-variable `PgConn` is set (env or json)
- call this program with the `migrate_shard` flag set to the url of the new camunda instance
```
./camunda-engine-wrapper -migrate_shard=http://shard_url:8080
```

## Vid Consistence Cleanup
- to clean inconsistent data this Program can be called with the `vid_cleanup` flag
- all config variables (except `ServerPort` and `LogLevel` ar used)
```
./camunda-engine-wrapper -vid_cleanup
```

## Docker
- env variables can be passed with the `-e` flag
- program-flags can be passed at the end of the docker run command 

```
docker build -t enginewrapper .
docker run -e PG_CONN=postgres://usr:pw@databasip:5432/shards?sslmode=disable enginewrapper -migrate_shard=http://new_shard_url:8080
```