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
go run . --mode=consumer
```

## Configuration

- `--mode`: `produce` or `consume`, no default
- `--bootstrap-servers`: default `localhost:9093`, Kafka bootstrap server comma delimited list
- `--consumer-group-id`: default `karapace-demo-<ts>`, Kafka consumer group.id name
- `--topic`: default `karapace-demo-topic`, topic name
- `--auto-offset-reset`: default `latest`, Kafka consumer auto.offset.reset
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
curl --silent -X GET http://localhost:8081/subjects | jq '.'
```

```json
[
  "karapace-demo-topic-key",
  "karapace-demo-topic-value"
]
```

Get versions of a subject:

```sh
curl --silent -X GET http://localhost:8081/subjects/karapace-demo-topic-value/versions | jq '.'
```

```json
[
  1
]
```

Get specific subject version:

```sh
curl --silent -X GET http://localhost:8081/subjects/karapace-demo-topic-value/versions/1 | jq '.'
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
curl --silent -X GET http://localhost:8081/schemas/ids/2 | jq '.'
```

```json
{
  "schema": "{\"name\":\"test\",\"type\":\"record\",\"fields\":[{\"name\":\"val\",\"type\":\"int\"}]}"
}
```
