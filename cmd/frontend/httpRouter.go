package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	machinery "github.com/RichardKnop/machinery/v1"
	"github.com/RichardKnop/machinery/v1/tasks"
	"github.com/gin-gonic/gin"
	"github.com/maystery/cqueue/pkg/common"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
)

type myForm struct {
	Image string   `form:"image"`
	Cmd   []string `form:"cmd"`
	Type  string   `form:"type"`
	Start string   `form:"start"`
	Stop  string   `form:"stop"`
}

func push(machineryServer *machinery.Server, task string, batch bool) (string, bool) {
	signature := &tasks.Signature{
		Args: []tasks.Arg{
			{
				Type:  "string",
				Value: task,
			},
		},
	}
	if batch {
		signature.Name = "run_local"
	} else {
		signature.Name = "run_docker"
	}
	asyncResult, err := machineryServer.SendTask(signature)
	if err != nil {
		log.Printf("Could not register new task: %s", err.Error())
		return "Could not register new task", false
	}
	return asyncResult.Signature.UUID, true
}

func indexHandler(c *gin.Context) {
	c.HTML(200, "form.html", nil)
}

func formHandler(c *gin.Context) {
	var cqueueForm myForm
	c.Bind(&cqueueForm)

	url := "http://localhost:8080/task"
	if cqueueForm.Type == "normal" {
		cqueueForm.Type = ""
	}
	cqueueForm.Cmd = strings.Split(cqueueForm.Cmd[0], " ")
	//cqueueForm.Cmd = cqueueForm.Cmd[1].Split(' ')
	jsonStr, _ := json.Marshal(cqueueForm)
	//var jsonStr = []byte()
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Error(err)
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	c.JSON(200, gin.H{"response": string(body)})
}

// NewHTTPRouter : Initiate new HTTP router
func NewHTTPRouter(machineryServer *machinery.Server) (router *gin.Engine) {

	router = gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.LoadHTMLGlob("views/*")

	router.GET("/", indexHandler)
	router.POST("/", formHandler)

	//Push new task, e.g curl -H 'Content-Type: application/json' -X POST -d'{"image":"ubuntu", "cmd":["sleep", "30"], "env":["foo=bar","foo2=bar2"]}' http://localhost:8080/task
	router.POST("/task", func(c *gin.Context) {
		var task common.Task
		if c.BindJSON(&task) == nil {
			// Marshal to string after validation
			taskStr, err := json.Marshal(task)
			if err != nil {
				log.Fatal(err)
			}
			var resp string
			var ok bool
			if strings.EqualFold(task.Type, "batch") {
				log.Printf("batch: true")
				resp, ok = push(machineryServer, string(taskStr), true)
			} else {
				log.Printf("batch: false")
				resp, ok = push(machineryServer, string(taskStr), false)
			}
			if ok {
				c.JSON(http.StatusOK, gin.H{"id": resp})
				log.Printf("%v - New task %s registered", time.Now().Format(time.StampMilli), taskStr)
			} else {
				c.JSON(http.StatusPreconditionFailed, gin.H{"status": resp})
				log.Printf("%v - Error! Response: %s", time.Now().Format(time.StampMilli), resp)
			}
		}
	})

	// Get status, e.g curl http://localhost:8080/task/$taskID
	router.GET("/task/:id", func(c *gin.Context) {
		id := c.Param("id")
		backend := machineryServer.GetBackend()
		state, err := backend.GetState(id)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"status": "Task not found"})
			log.Printf("%v - Task %s not found", time.Now().Format(time.StampMilli), id)
			return
		}
		c.JSON(http.StatusFound, gin.H{"status": state.State})
		log.Printf("%v - Task %s state: %s", time.Now().Format(time.StampMilli), id, state.State)

	})

	// Get result, e.g curl http://localhost:8080/task/$taskID/result
	router.GET("/task/:id/*keyword", func(c *gin.Context) {
		id := c.Param("id")
		backend := machineryServer.GetBackend()
		state, err := backend.GetState(id)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
			log.Printf("%v - Task %s not found", time.Now().Format(time.StampMilli), id)
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
