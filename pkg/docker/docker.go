package docker

import (
	"context"
	"io"
	"io/ioutil"

	"github.com/maystery/cqueue/pkg/common"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

type Docker struct {
	ctx context.Context
	cli *client.Client
}

func NewDockerCLI() (dockerCli *Docker, err error) {
	// cli, err := client.NewEnvClient()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	dockerCli = &Docker{
		ctx: context.Background(),
		cli: cli,
	}
	return
}

// TODO: Add authenticated docker pull
func (docker *Docker) ImagePull(image string) (out io.ReadCloser, err error) {
	out, err = docker.cli.ImagePull(docker.ctx, image, types.ImagePullOptions{})
	return
}

func (docker *Docker) ContainerLogs(container string) (out []byte, err error) {
	logReader, err := docker.cli.ContainerLogs(docker.ctx, container, types.ContainerLogsOptions{ShowStdout: true, ShowStderr: true})
	if err != nil {
		return
	}
	out, err = ioutil.ReadAll(logReader)
	return
}

func (docker *Docker) ContainerLaunch(task common.Task) (id string, err error) {
	containerConfig := &container.Config{
		Image: task.Image,
		Cmd:   task.Cmd,
		Env:   task.Env,
	}

	resp, err := docker.cli.ContainerCreate(docker.ctx, containerConfig, nil, nil, nil, task.ContainerName)
	if err != nil {
		return "", err
	}

	if err := docker.cli.ContainerStart(docker.ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		return "", err
	}

	statusCh, errCh := docker.cli.ContainerWait(docker.ctx, resp.ID, container.WaitConditionNotRunning)
	select {
	case <-statusCh:
		return resp.ID, nil
	case <-errCh:
		return "", <-errCh
	}
}

func (docker *Docker) ContainerRemove(containerID string) (err error) {
	err = docker.cli.ContainerRemove(docker.ctx, containerID, types.ContainerRemoveOptions{})
	return
}
