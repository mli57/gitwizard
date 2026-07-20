package docker

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// one published port from `docker compose ps --format json`. Field names have to
// match the JSON's, that's how the decoder knows what goes where.
type publisher struct {
	PublishedPort int
	Protocol      string
}

// list of ports for each container
type composeContainer struct {
	Publishers []publisher
}

// cd into repo path and runs the compose.yml to start the containers
// returns the host ports compose published and the project name (compose runs several containers, so there's no single container ID to hand back).
func RunCompose(repoPath string) ([]int, string, error) {
	// pass the name explicitly to stop docker from deriving its own name
	// this way we won't have to look for the name when running `compose down`
	projectName := strings.ToLower(filepath.Base(repoPath))

	cmd := exec.Command("docker", "compose", "-p", projectName, "up", "-d")
	cmd.Dir = repoPath // cd part
	cmd.Stdout = os.Stdout // print docker progress live
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return nil, "", fmt.Errorf("compose up failed: %w", err)
	}

	ports, err := composePorts(repoPath, projectName)
	if err != nil {
		return nil, "", err
	}
	return ports, projectName, nil
}

// Asks compose which host ports can actually be connnected
// Containers that don't map a port report a published port of 0 and get skipped
func composePorts(repoPath, projectName string) ([]int, error) {
	cmd := exec.Command("docker", "compose", "-p", projectName, "ps", "--format", "json")
	cmd.Dir = repoPath
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("could not list compose containers: %w", err)
	}

	seenPort := map[int]bool{} // each port is published twice, once for IPv4 and once for IPv6
	var ports []int

	// one JSON object per line, not a list of them, so decode until we run out
	decoder := json.NewDecoder(bytes.NewReader(output))
	for decoder.More() { // keep looping as long as there is unread JSON values
		var container composeContainer
		err := decoder.Decode(&container)
		if err != nil {
			return nil, fmt.Errorf("could not read compose containers: %w", err)
		}
		for _, published := range container.Publishers {
			// udp ports can't serve a web page, and a published port of 0 means
			// the container kept it to itself
			if published.Protocol != "tcp" || published.PublishedPort == 0 || seenPort[published.PublishedPort] {
				continue
			}
			seenPort[published.PublishedPort] = true
			ports = append(ports, published.PublishedPort)
		}
	}

	if len(ports) == 0 {
		return nil, fmt.Errorf("compose started, but no running service published a port to connect to (a service may have crashed - try `docker compose ps -a` in %s)", repoPath)
	}
	return ports, nil
}