package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	machinery "github.com/RichardKnop/machinery/v1"
	"github.com/RichardKnop/machinery/v1/tasks"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"github.com/maystery/cqueue/pkg/common"
)

func push(machineryServer *machinery.Server, task string) (string, bool) {
	signature := &tasks.Signature{
		Name: "run_docker",
		Args: []tasks.Arg{
			{
				Type:  "string",
				Value: task,
			},
		},
	}
	asyncResult, err := machineryServer.SendTask(signature)
	if err != nil {
		log.Debugf("Could not register new task: %s", err.Error())
		return "Could not register new task", false
	}
	return asyncResult.Signature.UUID, true
}

// NewHTTPRouter : Initiate new HTTP router
func NewHTTPRouter(machineryServer *machinery.Server) (router *gin.Engine) {

	router = gin.New()
	router.Use(gin.Recovery())

	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "Task failed"})
	})

	//Push new task, e.g curl -H 'Content-Type: application/json' -X POST -d'{"image":"ubuntu", "cmd":["sleep", "30"], "env":["foo=bar","foo2=bar2"]}' http://localhost:8080/task
	router.POST("/task", func(c *gin.Context) {
		var task common.Task
		if c.BindJSON(&task) == nil {
			// Marshal to string after validation
			taskStr, err := json.Marshal(task)
			if err != nil {
				log.Fatal(err)
			}
			resp, ok := push(machineryServer, string(taskStr))
			if ok {
				c.JSON(http.StatusOK, gin.H{"id": resp})
				log.Debugf("New task %s registered", taskStr)
			} else {
				c.JSON(http.StatusPreconditionFailed, gin.H{"status": resp})
				log.Debugf("Error! Response: %s", resp)
			}
		}
	})

	// Get status, e.g curl http://localhost:8080/task/$taskID
	router.GET("/task/:id", func(c *gin.Context) {
		id := c.Param("id")
		backend := machineryServer.GetBackend()
		state, err := backend.GetState(id)
		if err != nil {
			log.Debug("Task %s not found", id)
			c.JSON(http.StatusNotFound, gin.H{"status": "Task not found"})
			return
		}
		log.Debug("Task %s not found", id)
		c.JSON(http.StatusFound, gin.H{"status": state.State})
	})

	// Get result, e.g curl http://localhost:8080/task/$taskID/result
	router.GET("/task/:id/*keyword", func(c *gin.Context) {
		id := c.Param("id")
		backend := machineryServer.GetBackend()
		state, err := backend.GetState(id)
		if err != nil {
			log.Debug("Task %s not found", id)
			c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
			return
		}

		switch keyword := c.Param("keyword"); keyword {
		case "/result":
			if !state.IsCompleted() {
				c.JSON(http.StatusPreconditionFailed, gin.H{"error": "Task not completed yet"})
				return
			}
			if state.IsFailure() {
				c.JSON(http.StatusNotFound, gin.H{"error": "Task failed"})
				return
			}
			c.String(http.StatusOK, fmt.Sprintf("%v", state.Results[0].Value))
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": "Unknown keyword"})
			log.Debug("Unknown keyword: %s", keyword)
		}
	})

	// Purge task, e.g curl -X DELETE http://localhost:8080/task/$taskID
	router.DELETE("/task/:id", func(c *gin.Context) {
		id := c.Param("id")
		backend := machineryServer.GetBackend()
		_, err := backend.GetState(id)
		if err != nil {
			log.Debug("Task %s not found", id)
			c.JSON(http.StatusNotFound, gin.H{"status": "Task not found"})
			return
		}
		log.Debug("Task %s not found", id)

		err = backend.PurgeState(id)
		if err != nil {
			log.Debugf("Couldn't purge %s because of err: %v", id, err)
		} else {
			log.Debugf("Task %s purged", id)
			c.JSON(http.StatusOK, gin.H{"status": "Task purged"})
		}
	})

	handler := promhttp.Handler()
	router.GET("/metrics", func(c *gin.Context) {
		handler.ServeHTTP(c.Writer, c.Request)
	})

	return
}
