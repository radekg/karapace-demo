module github.com/radekg/karapace-demo

go 1.16

replace github.com/mycujoo/go-kafka-avro/v2 => github.com/radekg/go-kafka-avro/v2 v2.0.0-rg-ext.1

require (
	github.com/caarlos0/env/v6 v6.9.1 // indirect
	github.com/confluentinc/confluent-kafka-go v1.8.2
	github.com/hamba/avro v1.6.6 // indirect
	github.com/hashicorp/go-hclog v1.2.0
	github.com/mycujoo/go-kafka-avro/v2 v2.0.0
	github.com/open-policy-agent/opa v0.39.0
)
