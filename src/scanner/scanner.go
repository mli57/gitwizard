/*
Reads the pulled repo, determine its language and framework, and if it already has dockerfiles
*/

package scanner
import (
	"os"
	"path/filepath"
	"strings"
)

type RepoAnalysis struct {
	Language string
	Framework string
	RunMethod string
	Port int
	RuntimeVersion string
}

func Scan(repoPath string) (RepoAnalysis, error) {
	lang := detectLanguage(repoPath)
	return RepoAnalysis{
		Language: lang,
		RunMethod: detectRunMethod(repoPath),
		RuntimeVersion: detectRuntimeVersion(repoPath, lang),
	}, nil
}

func detectRuntimeVersion(repoPath, lang string) string {
	switch lang {
	case "python":
		// get version number from .python-version (pyenv)
		if b, err := os.ReadFile(filepath.Join(repoPath, ".python-version")); err == nil {
			return strings.TrimSpace(string(b))
		}
		// get version number from runtime.txt (e.g. "python-3.10.0")
		if b, err := os.ReadFile(filepath.Join(repoPath, "runtime.txt")); err == nil {
			v := strings.TrimSpace(string(b))
			return strings.TrimPrefix(v, "python-")
		}
	case "node":
		// get version number from .nvmrc
		if b, err := os.ReadFile(filepath.Join(repoPath, ".nvmrc")); err == nil {
			return strings.TrimSpace(string(b))
		}
	}
	return ""
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