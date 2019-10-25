package main

import (
	"flag"
	"net"
	"net/http"

	machinery "github.com/RichardKnop/machinery/v1"
	"github.com/RichardKnop/machinery/v1/config"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"gitlab.com/lpds-public/cqueue/tasks"
)

func main() {

	var (
		HTTPAddr   = flag.String("http", "0.0.0.0:8080", "address to listen for HTTP requests on")
		configPath = flag.String("config", "", "Path of config file")
	)

	flag.Parse()

	var (
		machineryConfig *config.Config
		err             error
	)
	if *configPath != "" {
		machineryConfig, err = config.NewFromYaml(*configPath, true)
	} else {
		machineryConfig, err = config.NewFromEnvironment(true)
	}
	if err != nil {
		log.Fatal(err)
	}

	// Create server instance
	machineryServer, err := machinery.NewServer(machineryConfig)
	if err != nil {
		log.Fatal(err)
	}

	// Register tasks
	tasks := map[string]interface{}{
		"run_docker": tasks.RunDocker,
	}

	err = machineryServer.RegisterTasks(tasks)
	if err != nil {
		log.Fatal(err)
	}

	// Check host and port validity
	_, _, err = net.SplitHostPort(*HTTPAddr)
	if err != nil {
		log.Fatal(err)
	}

	// Launch HTTP server
	log.Infof("HTTP server listening on %s", *HTTPAddr)

	var router *gin.Engine
	router = NewHTTPRouter(machineryServer)

	log.Fatal(http.ListenAndServe(*HTTPAddr, router))
}
