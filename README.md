## Configuration


| config.json              | env                      | desc                                                                                                                      |
|--------------------------|--------------------------|---------------------------------------------------------------------------------------------------------------------------|
| server_port                | SERVER_PORT               | port of wrapper api                                                                                             |
| zookeeper_url              | ZOOKEEPER_URL             | url to zookeeper                                                                                                          |
| kafka_group                | KAFKA_GROUP               | used kafka consumer group                                                                                        |
| deployment_topic           | DEPLOYMENT_TOPIC          | kafka topic of deployments                                                                                        |
| incident_topic             | INCIDENT_TOPIC            | kafka topic of incidents                                                                                                                          |
| wrapper_db                 | WRAPPER_DB                | connection string to postgres database to store virtual ids (e.g. postgres://usr:pw@databasip:5432/shards?sslmode=disable)                                                                                                                         |
| sharding_sb                | SHARDING_DB               | connection string to postgres database to store sharding information (e.g. postgres://usr:pw@databasip:5432/shards?sslmode=disable)                                                                                                                         |
| debug                      | DEBUG                     | more logs                                                      |

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