package config

import (
	"gopkg.in/yaml.v3"
	"log"
	"os"
	"time"
)

type VerifierConfig struct {
	RateLimit int           `yaml:"rate_limit"`
	RateTime  time.Duration `yaml:"rate_time"`
	Keys      []string      `yaml:"keys"`
}

func LoadVerifierConfiger() *VerifierConfig {
	yamlFile, err := os.ReadFile("verifier_config.yaml")
	if err != nil {
		log.Fatalf("failed to read config.yaml: %v", err)
	}

	var cfg VerifierConfig

	err = yaml.Unmarshal(yamlFile, &cfg)
	if err != nil {
		log.Fatalf("unmarshal failed with error: %v", err)
	}

	return &cfg
}
