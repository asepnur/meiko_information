package main

import (
	"flag"
	"log"
	"runtime"

	"github.com/asepnur/meiko_information/src/util/auth"
	"github.com/asepnur/meiko_information/src/util/conn"
	"github.com/asepnur/meiko_information/src/util/env"
	"github.com/asepnur/meiko_information/src/util/jsonconfig"
	"github.com/asepnur/meiko_information/src/webserver"
)

type configuration struct {
	Database  conn.DatabaseConfig `json:"database"`
	Webserver webserver.Config    `json:"webserver"`
	Auth      auth.Config         `json:"auth"`
}

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}

func main() {
	flag.Parse()

	// load configuration
	cfgenv := env.Get()
	config := &configuration{}
	isLoaded := jsonconfig.Load(&config, "/etc/meiko", cfgenv) || jsonconfig.Load(&config, "./files/etc/meiko", cfgenv)
	if !isLoaded {
		log.Fatal("Failed to load configuration")
	}

	// initiate instance
	conn.InitDB(config.Database)
	auth.Init(config.Auth)
	webserver.Start(config.Webserver)
}
