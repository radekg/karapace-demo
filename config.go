package main

const defaultLogLevel = "info"

type demoConfig struct {
	mode              string
	autoOffsetReset   string
	bootstrapServers  string
	consumerGroupId   string
	noAutoCommit      bool
	produceIntervalMs int64
	schemaURL         string
	topic             string
	logAsJSON         bool
	logLevel          string
}
