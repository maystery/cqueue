package tasks

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"os/exec"
	"strconv"
	"strings"
	"text/template"

	"github.com/maystery/cqueue/pkg/common"
	"github.com/maystery/cqueue/pkg/docker"
)

func RunLocal(arg string) (result string, err error) {
	var t common.Task
	err = json.Unmarshal([]byte(arg), &t)
	if err != nil {
		return
	}

	// Hack batch executor
	if strings.EqualFold(t.Type, "batch") {
		start, _ := strconv.Atoi(t.Start)
		stop, _ := strconv.Atoi(t.Stop)
		result = "\nBatch mode:\n" + "Start: " + t.Start + "\nStop: " + t.Stop + "\n\n"

		index := -1
		for i, v := range t.Cmd {
			if strings.Contains(v, "{{.}}") {
				index = i
			}
		}

		// There is no index field
		if index == -1 {
			result = "ERROR: There is no index field"
			return
		}
		for iii := start; iii < stop+1; iii++ {
			buf := new(bytes.Buffer)
			temp := make([]string, len(t.Cmd))
			copy(temp, t.Cmd)
			te := template.Must(template.New("").Parse(temp[index]))
			te.Execute(buf, iii)
			temp[index] = buf.String()
			cmd := exec.Command(t.Cmd[0])
			cmd.Args = temp
			cmd.Env = t.Env
			var output bytes.Buffer
			cmd.Stdout = &output
			err = cmd.Run()
			result += strconv.Itoa(iii) + ": " + output.String()
		}
		result = result[:len(result)-1]
		return
	}
	return
}

func RunDocker(arg string) (result string, err error) {
	var t common.Task
	err = json.Unmarshal([]byte(arg), &t)
	if err != nil {
		return
	}

	// Hack local executor
	if t.Type == "local" {
		cmd := exec.Command(t.Cmd[0])
		cmd.Args = t.Cmd
		cmd.Env = t.Env
		var output bytes.Buffer
		cmd.Stdout = &output
		err = cmd.Run()
		result = output.String()
		result = result[:len(result)-1]
		return
	}

	cli, err := docker.NewDockerCLI()
	if err != nil {
		return "", err
	}

	// Pull new image if available
	out, err := cli.ImagePull(t.Image)
	if err != nil {
		return "", err
	}
	defer out.Close()
	io.Copy(ioutil.Discard, out)

	id, err := cli.ContainerLaunch(t)
	if err != nil {
		return "", err
	}
	// TODO: investigate with error
	defer cli.ContainerRemove(id)

	// TODO: handle execution error
	output, err := cli.ContainerLogs(id)
	if err != nil {
		return "", err
	}
	result = string(output[8 : len(output)-1])
	return
}
