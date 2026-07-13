/*
Builds and runs Docker images from a repo path
*/
package docker

import (
	"archive/tar"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"net"

	"github.com/docker/docker/api/types/build"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

func Build(repoPath, imageName string) error {
	ctx := context.Background()

	// connect to Docker Desktop running on this machine
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return fmt.Errorf("could not connect to Docker: %w", err)
	}
	defer cli.Close()

	// Docker needs the repo files as a tar archive (the "build context")
	tar, err := tarDir(repoPath)
	if err != nil {
		return fmt.Errorf("could not create build context: %w", err)
	}

	// starts the image build, same as running "docker build"
	resp, err := cli.ImageBuild(ctx, tar, build.ImageBuildOptions{
		Tags:   []string{imageName},
		Remove: true, // clean up intermediate containers when done
	})
	if err != nil {
		return fmt.Errorf("build failed: %w", err)
	}
	defer resp.Body.Close()

	// Docker streams back JSON messages as the build progresses, one per step
	decoder := json.NewDecoder(resp.Body)
	for {
		var msg struct {
			Stream string `json:"stream"`
			Error  string `json:"error"`
		}
		if err := decoder.Decode(&msg); err == io.EOF {
			break
		} else if err != nil {
			return err
		}
		if msg.Error != "" {
			return fmt.Errorf("build error: %s", msg.Error)
		}
		if msg.Stream != "" {
			fmt.Print(msg.Stream)
		}
	}
	return nil
}

// packages the repo folder into a tar archive so Docker can use it as a build context
func tarDir(srcPath string) (io.Reader, error) {
	buf := new(bytes.Buffer)
	tw := tar.NewWriter(buf)

	err := filepath.Walk(srcPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil // skip directories, only add files
		}
		// get the file path relative to the repo root so Docker sees the right structure
		rel, err := filepath.Rel(srcPath, path)
		if err != nil {
			return err
		}
		// write the file metadata (name, size, permissions) into the tar
		hdr := &tar.Header{
			Name: filepath.ToSlash(rel),
			Mode: int64(info.Mode()),
			Size: info.Size(),
		}
		if err := tw.WriteHeader(hdr); err != nil {
			return err
		}
		// then write the file contents
		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()
		_, err = io.Copy(tw, f)
		return err
	})
	if err != nil {
		return nil, err
	}
	tw.Close()
	return buf, nil
}

// find an available on the local device to map to the container
func findFreePort() (int, error) {
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		return 0, err
	}
	listener.Close()
	return listener.Addr().(*net.TCPAddr).Port, nil
}

// Starts a container from imageName, mapping its internal port to a free host port.
// Returns the host port, the container ID (for the state store), and any error.
func Run(imageName string, port int) (int, string, error) {
	foundPort, err := findFreePort()
	if err != nil {
		return 0, "", err
	}

	hostPort := fmt.Sprintf("%d", foundPort)
	containerPort := fmt.Sprintf("%d/tcp", port)

	portBindings := nat.PortMap{
		nat.Port(containerPort): []nat.PortBinding{
			{HostIP: "127.0.0.1", HostPort: hostPort},
		},
	}

	// connect to Docker Desktop
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return 0, "", fmt.Errorf("could not connect to Docker: %w", err)
	}
	defer cli.Close()

	// create the container which tells Docker which image to use and how to map the internal port to the free port
	resp, err := cli.ContainerCreate(ctx,
		&container.Config{
			Image: imageName,
			ExposedPorts: nat.PortSet{
				nat.Port(containerPort): struct{}{},
			},
		},
		&container.HostConfig{
			PortBindings: portBindings,
		},
		nil, nil, "",
	)

	if err != nil {
		return 0, "", fmt.Errorf("could not create container: %w", err)
	}

	// start the container
	if err := cli.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		return 0, "", fmt.Errorf("could not start container: %w", err)
	}

	return foundPort, resp.ID, nil
}