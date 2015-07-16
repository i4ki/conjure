package conjure

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"time"

	"code.google.com/p/go-uuid/uuid"
	"github.com/fsouza/go-dockerclient"
)

// RabbitMQ is the entity that conjure client can spawn.
type RabbitMQ struct {
	*docker.Container
	client *Client
}

// New constructs a new rabbitmq container
func NewRabbitMQ(client *Client) *RabbitMQ {
	r := new(RabbitMQ)
	r.client = client

	return r
}

// Create a new rabbitmq container
func (ctn *RabbitMQ) Create() error {
	opts := docker.CreateContainerOptions{
		Name: "conjured-rabbitmq-" + uuid.New()[0:6],
		Config: &docker.Config{
			Image: "rabbitmq",
			ExposedPorts: map[docker.Port]struct{}{
				"5672/tcp": {},
			},
		},
		HostConfig: &docker.HostConfig{
			PortBindings: map[docker.Port][]docker.PortBinding{
				"5672/tcp": []docker.PortBinding{docker.PortBinding{HostPort: "35672"}}},
			PublishAllPorts: true,
			Privileged:      false,
		},
	}

	dockerCtn, err := ctn.client.CreateContainer(opts)

	ctn.Container = dockerCtn

	return err
}

// Run execute a container with rabbitmq daemon
// If docker fail to create the container, then try to pull the
// last version of rabbitmq image from registry. If fail again,
// then returns error.
// When rabbitmq is successfully started, block until it finish the
// setup and is ready to accept connections.
func (ctn *RabbitMQ) Run() error {
	err := ctn.Create()

	if err != nil {
		err = ctn.Pull()

		if err != nil {
			return err
		}

		err = ctn.Create()

		if err != nil {
			return err
		}
	}

	err = ctn.Start()

	if err != nil {
		return err
	}

	return nil
}

func (ctn *RabbitMQ) Pull() error {
	pullOpts := docker.PullImageOptions{
		Repository:   "rabbitmq",
		Tag:          "latest",
		OutputStream: os.Stdout,
	}

	err := ctn.client.PullImage(pullOpts, docker.AuthConfiguration{})

	if err != nil {
		return err
	}

	return nil
}

// Start the given container
func (ctn *RabbitMQ) Start() error {
	return ctn.client.StartContainer(ctn.ID, nil)
}

func (ctn *RabbitMQ) Stop() error {
	return ctn.client.StopContainer(ctn.ID, 3)
}

// WaitOK blocks until rabbitmq can accept connections on
// <ctn ip address>:5672
func (ctn *RabbitMQ) WaitOK(host string) error {
dial:
	conn, err := net.Dial("tcp", host+":35672")
	if err != nil {
		time.Sleep(500 * time.Millisecond)
		goto dial
	}

	fmt.Fprintf(conn, "AMQP%00091")
	_, err = bufio.NewReader(conn).ReadString('\n')

	if err != nil && err.Error() != "EOF" {
		//conn.Close()
		time.Sleep(500 * time.Millisecond)
		goto dial
	}

	return nil
}

func (ctn *RabbitMQ) Remove() error {
	rmCtnOpt := docker.RemoveContainerOptions{
		ID:            ctn.ID,
		RemoveVolumes: true,
		Force:         true,
	}

	return ctn.client.RemoveContainer(rmCtnOpt)
}
