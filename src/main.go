package main

import (
	"flag"
	"fmt"
	"go_proxy/proxy"
	"go_proxy/router"
	"os"
)

const (
	versionMsg = "v1.0.0"
	helpMsg    = `
	Bad proxy go is a primitive tool for breaching censorship
	Usage:
		bad_proxy <command> [arguments]
	The commands are:
		run                 start proxy according to config file
		build               build routing file from text files under rule-path to binary specified by router-path
	
	The arguments are:
		--help              show this help message
		--version           show version message
		--config            specify config file path, default conf/server.json
		--router-path       specify binary routing file path, default conf/rules.dat
		--rule-path         specify text rule files path, default ./rules
	`
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println(helpMsg)
		return
	}

	helpFlag := flag.Bool("help", false, "show help info")
	versionFlag := flag.Bool("version", false, "show version info")
	flag.Parse()

	if *versionFlag == true {
		fmt.Println("Bad Proxy Golang", versionMsg)
		return
	}
	if *helpFlag == true {
		fmt.Print(helpMsg)
		return
	}

	var err error
	switch os.Args[1] {
	case "run":
		err = run()
	case "build":
		err = build()
	default:
		fmt.Print(helpMsg)
	}
	if err != nil {
		fmt.Println(err)
	}
}

func run() (err error) {
	runCmd := flag.NewFlagSet("run", flag.ExitOnError)
	configPath := runCmd.String("config", "conf/server_config.json", "config file path")
	routerPath := runCmd.String("router-path", "conf/rules.dat", "router data path")
	err = runCmd.Parse(os.Args[2:])
	if err != nil {
		return
	}
	config, err := ReadConfig(*configPath)
	if err != nil {
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
	routerPath := flag.String("router-path", "conf/rules.dat", "router data path")
	err = buildCmd.Parse(os.Args[2:])
	if err != nil {
		return
	}
	err = router.WriteAllToGob(*rulePath, *routerPath)
	if err == nil {
		fmt.Printf("success build from %s to %s\n",
			*rulePath, *routerPath)
	}
	return
}
