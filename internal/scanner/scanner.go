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
	EntryPoint string
}

func Scan(repoPath string) (RepoAnalysis, error) {
	lang := detectLanguage(repoPath)
	framework := detectFramework(repoPath, lang)
	return RepoAnalysis{
		Language: lang,
		Framework: framework,
		RunMethod: detectRunMethod(repoPath),
		RuntimeVersion: detectRuntimeVersion(repoPath, lang),
		EntryPoint: detectEntryPoint(repoPath, framework),
	}, nil
}

// reads requirements.txt and returns the specified framework name
func detectFramework(repoPath, lang string) string {
	if lang != "python" {
		return ""
	}
	b, err := os.ReadFile(filepath.Join(repoPath, "requirements.txt"))
	if err != nil {
		return ""
	}
	content := strings.ToLower(string(b))
	switch {
	case strings.Contains(content, "flask"):
		return "flask"
	case strings.Contains(content, "fastapi"):
		return "fastapi"
	case strings.Contains(content, "django"):
		return "django"
	case strings.Contains(content, "streamlit"):
		return "streamlit"
	case strings.Contains(content, "gradio"):
		return "gradio"
	}
	return ""
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

// only the bare compose filenames are recognized (also docker-compose.override.yml,which is auto-merged). 
// Named variants of these files are deliberately left out.
var composeFilenames = []string{
	"docker-compose.yml",
	"docker-compose.yaml",
	"compose.yml",
	"compose.yaml",
}

func detectRunMethod(repoPath string) string {
	// return 'compose' if we find any recognized compose filename
	for _, name := range composeFilenames {
		if _, err := os.Stat(filepath.Join(repoPath, name)); err == nil {
			return "compose"
		}
	}

	// same for dockerfile
	_, err := os.Stat(filepath.Join(repoPath, "Dockerfile"))
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

func detectEntryPoint(repoPath string, framework string) string {
	if framework != "flask" && framework != "fastapi" {
		return ""
	}
	pattern := "Flask("
	if framework == "fastapi" {
		pattern = "FastAPI("
	}
	found := ""
	filepath.Walk(repoPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(path, ".py") {
			return nil
		}
		b, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		if strings.Contains(string(b), pattern) {
			found = filepath.Base(path)
		}
		return nil
	})
	return found
}