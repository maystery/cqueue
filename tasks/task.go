package tasks

import (
	"bytes"
	"encoding/json"
	"os/exec"

	"github.com/maystery/cqueue/pkg/common"
	"github.com/maystery/cqueue/pkg/docker"
)

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
	// ??
	// Discard is an io.Writer on which all Write calls succeed without doing anything.
	//io.Copy(ioutil.Discard, out)

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
	result = string(output)
	return
}
