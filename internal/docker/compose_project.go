package docker

import (
	"fmt"
	"os/exec"
	"path/filepath"
)

// cd into repo path and runs the compose.yml
func RunCompose(repoPath string) (int, string, error) {
	projectName := filepath.Base(repoPath)
	cmd := exec.Command("docker", "compose", "up", "-d")
	cmd.Dir = repoPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		return 0, "", fmt.Errorf("compose up failed: %w\n%s", err, output)
	}
	return 0, projectName, nil
}