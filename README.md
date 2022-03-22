# Karapace demo

A demo application using the Aiven [Karapace](https://karapace.io/) schema registry with golang Kafka client.

## Usage

Start the cluster:

```sh
make clean up
```

The cluster requires following ports available on the host:

- `2181`, `2182`, `2183`: ZooKeeper ports
- `9093`, `9094`, `9094`: Kafka ports
- `8081`: Karapace schema registry port

In another terminal, start the producer:

```sh
go run . --mode=produce
```

In yet another terminal, start the consumer:

```sh
go run . --mode=consume
```

## Configuration

- `--mode`: `produce` or `consume`, no default
- `--bootstrap-servers`: default `localhost:9093`, Kafka bootstrap server comma delimited list
- `--consumer-group-id`: default `karapace-demo-<ts>`, Kafka consumer group.id name
- `--topic`: default `karapace-demo-topic`, topic name
- `--auto-offset-reset`: default `latest`, Kafka consumer auto.offset.reset: `smallest`, `earliest`, `beginning`, `largest`, `latest`, `end`, `error`. `none`
- `--no-auto-commit`: default `false`, if set, disables the auto-commit of the consumer offset
- `--produce-interval-ms`: default `1000`, message produce interval
- `--schema-registry-url`: default `http://localhost:8081`, Karapace schema URL
- `--log-as-json`: default `false`, if set, log as JSON
- `--log-level`: default `info`, log level

## What does this do

In the `produce` mode, produces a message to the topic every `--produce-interval-ms`. A message is an instance of hard-coded `Val` type with a single `Val int` property. Message is serialized to Avro using the hard-coded schema.

In the `consume` mode, consumes messages as fast as possible from the topic. Each consumed message is deserialized using the schema.

## Validate the registry state

Assuming default settings:

```sh
curl --silent http://localhost:8081/subjects | jq '.'
```

```json
[
  "karapace-demo-topic-key",
  "karapace-demo-topic-value"
]
```

Get versions of a subject:

```sh
curl --silent http://localhost:8081/subjects/karapace-demo-topic-value/versions | jq '.'
```

```json
[
  1
]
```

Get specific subject version:

```sh
curl --silent http://localhost:8081/subjects/karapace-demo-topic-value/versions/1 | jq '.'
```

```json
{
  "id": 2,
  "schema": "{\"name\":\"test\",\"type\":\"record\",\"fields\":[{\"name\":\"val\",\"type\":\"int\"}]}",
  "subject": "karapace-demo-topic-value",
  "version": 1
}
```

Get schema by ID:

```sh
curl --silent http://localhost:8081/schemas/ids/2 | jq '.'
```

```json
{
  "schema": "{\"name\":\"test\",\"type\":\"record\",\"fields\":[{\"name\":\"val\",\"type\":\"int\"}]}"
}
```

### REST proxy

REST proxy runs in the `karapace-rest` container. This container is configured to run on port `8082`. To list topics:

```sh
curl --silent http://localhost:8081/topics | jq '.'
```

```json
[
  "_schemas",
  "karapace-demo-topic",
  "__consumer_offsets"
]
```

Get details of a single topic:

```sh
curl --silent -X GET http://localhost:8082/topics/karapace-demo-topic | jq '.'
```

```json
{
  "configs": {
    "cleanup.policy": "delete",
    "compression.type": "producer",
    "delete.retention.ms": "86400000",
    "file.delete.delay.ms": "60000",
    "flush.messages": "9223372036854775807",
    "flush.ms": "9223372036854775807",
    "follower.replication.throttled.replicas": "",
    "index.interval.bytes": "4096",
    "leader.replication.throttled.replicas": "",
    "max.compaction.lag.ms": "9223372036854775807",
    "max.message.bytes": "1048588",
    "message.downconversion.enable": "true",
    "message.format.version": "3.0-IV1",
    "message.timestamp.difference.max.ms": "9223372036854775807",
    "message.timestamp.type": "CreateTime",
    "min.cleanable.dirty.ratio": "0.5",
    "min.compaction.lag.ms": "0",
    "min.insync.replicas": "1",
    "preallocate": "false",
    "retention.bytes": "-1",
    "retention.ms": "604800000",
    "segment.bytes": "1073741824",
    "segment.index.bytes": "10485760",
    "segment.jitter.ms": "0",
    "segment.ms": "604800000",
    "unclean.leader.election.enable": "false"
  },
  "name": "karapace-demo-topic",
  "partitions": [
    {
      "leader": 0,
      "partition": 0,
      "replicas": [
        {
          "broker": 0,
          "in_sync": true,
          "leader": true
        }
      ]
    }
  ]
}
```

## Clean up

```sh
make clean
```
