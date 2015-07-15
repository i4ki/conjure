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

// PostGIS is the entity that conjure client can spawn.
type PostGIS struct {
	*docker.Container
	client *Client
}

// New constructs a new postgis container
func NewPostGIS(client *Client) *PostGIS {
	r := new(PostGIS)
	r.client = client

	return r
}

// Create a new postgis container
func (ctn *PostGIS) Create() error {
	opts := docker.CreateContainerOptions{
		Name: "conjured-postgis-" + uuid.New()[0:6],
		Config: &docker.Config{
			Image: "postgis",
			ExposedPorts: map[docker.Port]struct{}{
				"5432/tcp": {},
			},
		},
		HostConfig: &docker.HostConfig{
			PortBindings: map[docker.Port][]docker.PortBinding{
				"5432/tcp": []docker.PortBinding{docker.PortBinding{HostPort: "35432"}}},
			PublishAllPorts: true,
			Privileged:      false,
		},
	}

	dockerCtn, err := ctn.client.CreateContainer(opts)

	ctn.Container = dockerCtn

	return err
}

// Run execute a container with postgis daemon
// If docker fail to create the container, then try to pull the
// last version of postgis image from registry. If fail again,
// then returns error.
// When postgis is successfully started, block until it finish the
// setup and is ready to accept connections.
func (ctn *PostGIS) Run() error {
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

	// block until postgis can accept connections on port 35432
	return ctn.WaitOK()
}

func (ctn *PostGIS) Pull() error {
	pullOpts := docker.PullImageOptions{
		Repository:   "postgis",
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
func (ctn *PostGIS) Start() error {
	return ctn.client.StartContainer(ctn.ID, nil)
}

func (ctn *PostGIS) Stop() error {
	return ctn.client.StopContainer(ctn.ID, 3)
}

// WaitOK blocks until postgis can accept connections on
// <ctn ip address>:5432
func (ctn *PostGIS) WaitOK() error {
dial:
	conn, err := net.Dial("tcp", "localhost:35432")
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

func (ctn *PostGIS) Remove() error {
	rmCtnOpt := docker.RemoveContainerOptions{
		ID:            ctn.ID,
		RemoveVolumes: true,
		Force:         true,
	}

	return ctn.client.RemoveContainer(rmCtnOpt)
}
