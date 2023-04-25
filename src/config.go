package main

import (
	"encoding/json"
	"log"
	"os"
)

type InboundConfig struct {
	Host        string `json:"host"`
	Port        string `json:"port"`
	Protocol    string `json:"protocol"`
	Transmit    string `json:"transmit"`
	WsPath      string `json:"ws_path"`
	TlsCertPath string `json:"tls_cert_path"`
	TlsKeyPath  string `json:"tls_key_path"`
}

type OutboundConfig struct {
	Host     string `json:"host"`
	Port     string `json:"port"`
	Protocol string `json:"protocol"`
	Transmit string `json:"transmit"`
	WsPath   string `json:"ws_path"`
}

type Config struct {
	Inbound  []InboundConfig  `json:"inbound"`
	Outbound []OutboundConfig `json:"outbound"`
}

func ReadConfig(path string) (Config, error) {
	file, err := os.ReadFile(path)
	if err != nil {
		log.Printf("failed to read file ")
		return Config{}, nil
	}
	var config Config
	err = json.Unmarshal(file, &config)
	if err != nil {
		log.Printf("failed to unmarshal")
		return Config{}, nil
	}
	return config, nil
}
