package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"go_proxy/proxy"
	"go_proxy/router"
	"os"
)

const (
	versionMsg = "v1.0.3"
	helpMsg    = `
	Bad proxy go is a primitive tool for breaching censorship
	Run application:
		bad_proxy <options>
	Options:
		--help              show this help message
		--version           show version message
		--config            specify config file path, default conf/config.json
		--router-path       specify binary routing file path, default conf/rules.dat

	Build routing file from text to binary:
		bad_proxy build <options>
	Options:
		--rule-path         specify text rule files directory path, default ./rules
		--router-path       specify output binary routing file path, default conf/rules.dat
	`
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

func main() {
	helpFlag := flag.Bool("help", false, "show help info")
	versionFlag := flag.Bool("version", false, "show version info")
	flag.Parse()

	if len(os.Args) >= 2 && os.Args[1] == "build" {
		if err := build(); err != nil {
			fmt.Println(err)
			fmt.Print(helpMsg)
		}
		return
	}

	if err := run(); err != nil {
		fmt.Print("failed to run app")
	}

	if *versionFlag == true {
		fmt.Println("Bad Proxy Golang", versionMsg)
		return
	}
	if *helpFlag == true {
		fmt.Print(helpMsg)
		return
	}
}

func run() (err error) {
	configPath := flag.String("config", "config.json", "config file path")
	routerPath := flag.String("router-path", "rules.dat", "router data path")
	flag.Parse()

	config, err := ReadConfig(*configPath)
	if err != nil {
		fmt.Print("failed to read config\n")
		return
	}
	config.RouterPath = *routerPath
	proxy.NewProxy(config).Start()
	<-make(chan os.Signal)
	return
}

func build() (err error) {
	buildCmd := flag.NewFlagSet("build", flag.ExitOnError)
	rulePath := flag.String("rule-path", "./rules", "router data path")
	routerPath := flag.String("router-path", "rules.dat", "router data path")
	if err = buildCmd.Parse(os.Args[2:]); err != nil {
		fmt.Println(err)
		return
	}
	if err = router.WriteAllToGob(*rulePath, *routerPath); err == nil {
		fmt.Printf("success build from %s to %s\n", *rulePath, *routerPath)
	}
	return
}
