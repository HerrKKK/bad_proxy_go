package main

import (
	"encoding/json"
	"go_proxy/proxy"
	"os"
)

func ReadConfig(path string) (config proxy.Config, err error) {
	file, err := os.ReadFile(path)
	if err != nil {
		return
	}
	err = json.Unmarshal(file, &config)
	if err != nil {
		return
	}
	return config, nil
}
