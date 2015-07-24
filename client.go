package conjure

import (
	"encoding/json"
	"fmt"

	"github.com/fsouza/go-dockerclient"
)

type Entity struct {
	Name       string             `json:"Name"`
	Config     *docker.Config     `json:"Config"`
	HostConfig *docker.HostConfig `json:"HostConfig"`
}

type Client struct {
	*docker.Client
}

func NewClient() (*Client, error) {
	var (
		client Client
	)

	// Start the amqp backend on test startup
	endpoint := "unix:///var/run/docker.sock"
	dockerClient, err := docker.NewClient(endpoint)

	if err != nil {
		return nil, err
	}

	client = Client{dockerClient}

	return &client, err
}

func (c *Client) Run(containerSpec string) (*docker.Container, error) {
	container := Entity{}

	err := json.Unmarshal([]byte(containerSpec), &container)

	if err != nil {
		return nil, err
	}

	fmt.Printf("Spec: %+v\n", container)

	opts := docker.CreateContainerOptions{
		Name:       container.Name,
		Config:     container.Config,
		HostConfig: container.HostConfig,
	}

	dockerCtn, err := c.CreateContainer(opts)

	if err != nil {
		return nil, err
	}
	
	err = c.StartContainer(dockerCtn.ID, nil)
	
	if err != nil {
		return nil, err
	}

	return dockerCtn, nil
}

func (c *Client) Remove(id string) error {
	rmCtnOpt := docker.RemoveContainerOptions{
		ID:            id,
		RemoveVolumes: true,
		Force:         true,
	}

	return c.RemoveContainer(rmCtnOpt)
}
