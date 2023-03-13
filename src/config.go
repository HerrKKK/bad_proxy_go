package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type InboundConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Protocol string `json:"protocol"`
}

type OutboundConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Protocol string `json:"protocol"`
}

type Config struct {
	Inbound  InboundConfig  `json:"inbound"`
	Outbound OutboundConfig `json:"outbound"`
}

func ReadConfig(path string) (Config, error) {
	file, err := os.ReadFile(path)
	if err != nil {
		fmt.Printf("failed to read file ")
		return Config{}, nil
	}
	var config Config
	err = json.Unmarshal(file, &config)
	if err != nil {
		fmt.Printf("failed to unmarshal")
		return Config{}, nil
	}
	fmt.Printf("config Struct: %#v\n", config)
	return config, nil
}
