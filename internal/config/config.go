package config

import (
	"log"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Cassandra CassandraConfig `yaml:"cassandra"`
	Kafka     KafkaConfig     `yaml:"kafka"`
	GRPC      GRPCConfig      `yaml:"grpc"`
	Logger    LoggerConfig    `yaml:"logger"`
}

type CassandraConfig struct {
	Hosts    []string `yaml:"hosts"    env:"CASSANDRA_HOSTS"     env-separator:","`
	Keyspace string   `yaml:"keyspace" env:"CASSANDRA_KEYSPACE"`
	Username string   `yaml:"username" env:"CASSANDRA_USERNAME"`
	Password string   `yaml:"password" env:"CASSANDRA_PASSWORD"`
}

type KafkaConfig struct {
	Brokers []string `yaml:"brokers"        env:"KAFKA_BROKERS"  env-separator:","`
	Topic   string   `yaml:"topic"          env:"KAFKA_TOPIC"`
	GroupID string   `yaml:"group_id"       env:"KAFKA_GROUP_ID"`
}

type GRPCConfig struct {
	Port int `yaml:"port" env:"GRPC_PORT" env-default:"50051"`
}

type LoggerConfig struct {
	Env string `yaml:"env" env:"LOGGER_ENV" env-default:"development"`
}

func MustLoad() *Config {
	var cfg Config

	if err := cleanenv.ReadConfig("config.yaml", &cfg); err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	return &cfg
}
