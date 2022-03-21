package main

import (
	"context"
	"flag"
	"fmt"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/hashicorp/go-hclog"
	kafkaavro "github.com/mycujoo/go-kafka-avro/v2"
)

type Val struct {
	Val int `avro:"val"`
}

func main() {

	cfg := &demoConfig{}

	flag.StringVar(&cfg.mode, "mode", "", "produce or consume")
	flag.StringVar(&cfg.bootstrapServers, "bootstrap-servers", "localhost:9093", "Kafka bootstrap server comma delimited list")
	flag.StringVar(&cfg.consumerGroupId, "consumer-group-id", fmt.Sprintf("karapace-demo-%d", time.Now().Unix()), "Kafka consumer group.id name")
	flag.StringVar(&cfg.topic, "topic", "karapace-demo-topic", "Topic name")
	flag.StringVar(&cfg.autoOffsetReset, "auto-offset-reset", "latest", "Kafka consumer auto.offset.reset")
	flag.Int64Var(&cfg.produceIntervalMs, "produce-interval-ms", 1000, "Message produce interval")
	flag.StringVar(&cfg.schemaURL, "schema-registry-url", "http://localhost:8081", "Karapace schema URL")
	flag.BoolVar(&cfg.logAsJSON, "log-as-json", false, "log as JSON")
	flag.StringVar(&cfg.logLevel, "log-level", defaultLogLevel, "log level")

	flag.Parse()

	ctx, cancelFunc := context.WithCancel(context.Background())

	logger := hclog.New(&hclog.LoggerOptions{
		Name:       "karapace-demo",
		Level:      hclog.LevelFromString(cfg.logLevel),
		JSONFormat: cfg.logAsJSON,
	}).With("topic", cfg.topic)

	switch cfg.mode {
	case "produce":

		c := make(chan os.Signal)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		go func() {
			<-c
			logger.Info("shutting down...")
			cancelFunc()
		}()

		exitCode := runProduce(ctx, logger, cfg)

		os.Exit(exitCode)

	case "consume":

		c := make(chan os.Signal)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		go func() {
			<-c
			logger.Info("shutting down...")
			cancelFunc()
		}()

		cExitCode := make(chan int)
		go func() {
			cExitCode <- runConsume(ctx, logger, cfg)
		}()

		os.Exit(<-cExitCode)

	default:
		logger.Error("unknown mode", "mode", cfg.mode)
		os.Exit(1)
	}

}

func runProduce(ctx context.Context, logger hclog.Logger, cfg *demoConfig) int {

	schemaURL, err := url.Parse(cfg.schemaURL)
	if err != nil {
		logger.Error("failed parsing schema URL", "reason", err)
		return 1
	}

	producer, err := kafkaavro.NewProducer(
		cfg.topic,
		`"string"`,
		`{"type": "record", "name": "test", "fields" : [{"name": "val", "type": "int", "default": 0}]}`,
		kafkaavro.WithKafkaConfig(&kafka.ConfigMap{
			"bootstrap.servers": cfg.bootstrapServers,
		}),
		kafkaavro.WithSchemaRegistryURL(schemaURL),
	)

	if err != nil {
		logger.Error("failed creating producer", "reason", err)
		return 1
	}

	for {

		select {
		case <-ctx.Done():
			logger.Info("Shutting down...")
			return 0
		case <-time.After(time.Millisecond * time.Duration(cfg.produceIntervalMs)):
			ts := time.Now().Unix()
			key := fmt.Sprintf("key-%d", ts)
			if err := producer.Produce(key, &Val{Val: (int)(ts)}, nil); err != nil {
				logger.Error("failed producing a record", err)
			}
			logger.Info("Produced a message", "key", key)
		}
	}

}

func runConsume(ctx context.Context, logger hclog.Logger, cfg *demoConfig) int {

	schemaURL, err := url.Parse(cfg.schemaURL)
	if err != nil {
		logger.Error("failed parsing schema URL", "reason", err)
		return 1
	}

	c, err := kafkaavro.NewConsumer(
		[]string{cfg.topic},
		func(topic string) interface{} {
			return &Val{}
		},
		kafkaavro.WithKafkaConfig(&kafka.ConfigMap{
			"bootstrap.servers": cfg.bootstrapServers,
			"group.id":          cfg.consumerGroupId,
			// avro kafka client commits automatically anyway...
			//"enable.auto.commit":      false,
			"socket.keepalive.enable": true,
			"auto.offset.reset":       cfg.autoOffsetReset,
		}),
		kafkaavro.WithSchemaRegistryURL(schemaURL),
		kafkaavro.WithEventHandler(func(event kafka.Event) {
			logger.Debug("kafka event", "event", event)
		}),
	)

	if err != nil {
		logger.Error("failed creating consumer", "reason", err)
	}

	for {
		select {
		case <-ctx.Done():
			logger.Info("shutting down consumer...")
			return 0
		default:
			// reiterate
		}

		msg, err := c.ReadMessage(1000)
		if err != nil {
			logger.Error("error when consuming", "reason", err)
			continue
		}

		if msg == nil {
			continue
		}

		switch v := msg.Value.(type) {
		case *Val:
			logger.Info("consumed a record",
				"key", string(msg.Key),
				"partition", msg.TopicPartition.Partition,
				"offset", msg.TopicPartition.Offset,
				"value", v)
		}
	}

}
