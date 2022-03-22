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

const avroKeySchema = `"string"`
const avroValueSchema = `{
	"type": "record",
	"name": "test",
	"fields" : [
		{
			"name": "val",
			"type": "int",
			"default": 0
		}
	]
}`

type Val struct {
	Val int `avro:"val"`
}

func main() {

	cfg := &demoConfig{}

	flag.StringVar(&cfg.mode, "mode", "", "produce or consume")
	flag.StringVar(&cfg.bootstrapServers, "bootstrap-servers", "localhost:9093", "Kafka bootstrap server comma delimited list")
	flag.StringVar(&cfg.consumerGroupId, "consumer-group-id", fmt.Sprintf("karapace-demo-%d", time.Now().Unix()), "Kafka consumer group.id name")
	flag.StringVar(&cfg.topic, "topic", "karapace-demo-topic", "Topic name")
	flag.StringVar(&cfg.autoOffsetReset, "auto-offset-reset", "latest", "Kafka consumer auto.offset.reset: smallest, earliest, beginning, largest, latest, end, error. none")
	flag.BoolVar(&cfg.noAutoCommit, "no-auto-commit", false, "If set, disables the auto-commit of the consumer offset")
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
			cancelFunc()
		}()

		cExitCode := make(chan int)
		go func() {
			cExitCode <- runProduce(ctx, logger, cfg)
		}()

		os.Exit(<-cExitCode)

	case "consume":

		c := make(chan os.Signal)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		go func() {
			<-c
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
		avroKeySchema,
		avroValueSchema,
		kafkaavro.WithKafkaConfig(&kafka.ConfigMap{
			"bootstrap.servers": cfg.bootstrapServers,
		}),
		kafkaavro.WithSchemaRegistryURL(schemaURL),
	)

	if err != nil {
		logger.Error("failed creating producer", "reason", err)
		return 1
	}

	chanDelivery := make(chan kafka.Event)

	go func() {
		for {
			select {
			case event := <-chanDelivery:
				switch vEvent := event.(type) {
				case *kafka.Message:
					logger.Info("delivery notification",
						"key", string(vEvent.Key),
						"headers", vEvent.Headers,
						"partition", vEvent.TopicPartition)
				default:
					logger.Info("delivery notification",
						"event", event.String(),
						"event-type", fmt.Sprintf("%+T", event))
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	for {

		select {
		case <-ctx.Done():
			logger.Info("shutting down producer...")
			producer.Close()
			logger.Info("producer stopped")
			return 0
		case <-time.After(time.Millisecond * time.Duration(cfg.produceIntervalMs)):
			ts := time.Now().Unix()
			key := fmt.Sprintf("key-%d", ts)
			if err := producer.Produce(key, &Val{Val: (int)(ts)}, chanDelivery); err != nil {
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

	kafkaCfg := &kafka.ConfigMap{
		"bootstrap.servers":       cfg.bootstrapServers,
		"group.id":                cfg.consumerGroupId,
		"enable.auto.commit":      !cfg.noAutoCommit,
		"socket.keepalive.enable": true,
	}
	if cfg.autoOffsetReset != "none" {
		kafkaCfg.SetKey("auto.offset.reset", cfg.autoOffsetReset)
	}

	c, err := kafkaavro.NewConsumer(
		[]string{cfg.topic},
		func(topic string) interface{} {
			return &Val{}
		},
		kafkaavro.WithKafkaConfig(kafkaCfg),
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
			c.Close()
			logger.Info("consumer stopped")
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
