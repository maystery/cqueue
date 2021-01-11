package main

import (
	"flag"
	"math/rand"
	"time"

	"github.com/RichardKnop/machinery/v1"
	"github.com/RichardKnop/machinery/v1/config"
	tsk "github.com/maystery/cqueue/tasks"
	log "github.com/sirupsen/logrus"
)

var (
	configPath  = flag.String("config", "", "Path of config file")
	concurrency = flag.Int("concurrency", 1, "Number of concurrent workers")
	instanceTag = flag.Int("tag", -1, "Tag of the worker instance")
	timeout     = flag.Duration("timeout", 0, "Exit after timeout")
	batch       = flag.Bool("batch", false, "Batch worker")
)

func main() {

	flag.Parse()

	if *instanceTag == -1 {
		rand.Seed(time.Now().Unix())
		*instanceTag = rand.Int()
	}

	var (
		machineryConfig *config.Config
		err             error
	)
	if *configPath != "" {
		machineryConfig, err = config.NewFromYaml(*configPath, true)
	} else {
		machineryConfig, err = config.NewFromEnvironment()
		machineryConfig.TLSConfig = nil
	}
	if err != nil {
		log.Fatal(err)
	}

	// Create server instance
	machineryServer, err := machinery.NewServer(machineryConfig)
	if err != nil {
		log.Fatal(err)
	}

	var tasks map[string]interface{}
	if *batch {
		tasks = map[string]interface{}{
			"run_local": tsk.RunLocal,
		}
	} else {
		// Register tasks
		tasks = map[string]interface{}{
			"run_docker": tsk.RunDocker,
		}
	}

	err = machineryServer.RegisterTasks(tasks)
	if err != nil {
		log.Fatal(err)
	}

	// The second argument is a consumer tag
	// Ideally, each worker should have a unique tag (worker1, worker2 etc)
	worker := machineryServer.NewWorker("machinery_worker", *instanceTag)
	worker.Concurrency = *concurrency

	if *timeout > 0 {
		go func() {
			time.Sleep(*timeout)
			worker.Quit()
		}()
	}

	worker.Launch()
	if err != nil {
		log.Fatal(err)
	}
}
