package main

import (
	"encoding/json"
	"go_proxy/proxy"
	"log"
	"os"
)

type AppProtType string

const (
	HTTP AppProtType = "http"
	BTP  AppProtType = "btp"
)

func ReadConfig(path string) (config proxy.Config, err error) {
	file, err := os.ReadFile(path)
	if err != nil {
		log.Printf("failed to read file ")
		return config, nil
	}
	err = json.Unmarshal(file, &config)
	if err != nil {
		log.Printf("failed to unmarshal")
		return config, nil
	}
	return config, nil
}
