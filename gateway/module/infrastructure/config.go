package infrastructure

import (
	"github.com/joho/godotenv"
	"go-simpler.org/env"
)

type Config struct {
	DatabaseUrl           string `env:"DATABASE_URL"`
	DeviceEventKafkaTopic string `env:"DEVICE_EVENT_KAFKA_TOPIC"`
	GRPCPort              string `env:"GRPC_PORT"`
}

func NewConfig() (*Config, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, err
	}

	c := &Config{}
	if err := env.Load(c, nil); err != nil {
		return nil, err
	}

	return c, nil
}
