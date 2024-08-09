package main

import (
	"flag"
	"fmt"
	"go_proxy/proxy"
	"os"
)

const (
	versionMsg = "v1.0.4"
	helpMsg    = `
	Bad proxy go is a primitive tool for breaching censorship
	Run application:
		bad_proxy <options>
	Options:
		--help              show this help message
		--version           show version message
		--config            specify config file path, default conf/config.json
		--router-path       specify binary routing file path, default conf/rules.dat
	`
)

func main() {
	helpFlag := flag.Bool("help", false, "show help info")
	versionFlag := flag.Bool("version", false, "show version info")
	configPath := flag.String("config", "config.json", "config file path")
	routerPath := flag.String("router-path", "rules.dat", "router data path")
	flag.Parse()

	if *versionFlag == true {
		fmt.Println("Bad Proxy Golang", versionMsg)
		return
	}
	if *helpFlag == true {
		fmt.Print(helpMsg)
		return
	}

	err := proxy.Startup(*configPath, *routerPath)
	if err != nil {
		fmt.Println("Error starting proxy:", err)
		os.Exit(0)
	}
	<-make(chan os.Signal)
}
