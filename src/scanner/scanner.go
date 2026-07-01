package scanner
import (
	"os"
	"path/filepath"
)

type RepoAnalysis struct {
	Language string
	Framework string
	RunMethod string
	Port int
}

func Scan(repoPath string) (RepoAnalysis, error) {
	return RepoAnalysis{
		Language:  detectLanguage(repoPath),
		RunMethod: detectRunMethod(repoPath),
	}, nil
}

func detectRunMethod(repoPath string) string {
	// return 'compose' if we find the .yml
	_, err := os.Stat(filepath.Join(repoPath, "docker-compose.yml"))
	if err == nil {
		return "compose"
	}

	// same for dockerfile
	_, err = os.Stat(filepath.Join(repoPath, "Dockerfile"))
	if err == nil {
		return "dockerfile"
	}
	return "generate" // if we found neither, we have to search the repo for dependencies
}

func detectLanguage(repoPath string) string {
	_, err := os.Stat(filepath.Join(repoPath, "requirements.txt"))
	if err == nil {
		return "python"
	}

	_, err = os.Stat(filepath.Join(repoPath, "go.mod"))
	if err == nil {
		return "go"
	}

	_, err = os.Stat(filepath.Join(repoPath, "package.json"))
	if err == nil {
		return "node"
	}
	return "unknown"
}