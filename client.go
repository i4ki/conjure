package conjure

import "github.com/fsouza/go-dockerclient"

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
