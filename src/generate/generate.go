/*
Creates dockerfiles for repos that don't have them
*/
package generate

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mli57/gitwizard/src/scanner"
)

var dockerfileTemplates = map[string]string{
	"node": `FROM node:20-alpine
WORKDIR /app
COPY package*.json ./
RUN npm install
COPY . .
EXPOSE 3000
CMD ["npm", "start"]
`,

	"python": `FROM python:3.11-slim
WORKDIR /app
COPY requirements.txt .
RUN pip install -r requirements.txt
COPY . .
EXPOSE 5000
CMD ["python", "main.py"]
`,

	"go": `FROM golang:1.22-alpine
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o app .
EXPOSE 8080
CMD ["./app"]
`,
}

func Dockerfile(repoPath string, analysis scanner.RepoAnalysis) error {
	switch analysis.RunMethod {
	case "compose", "dockerfile":
		return nil // already have docker assets, nothing to generate
	case "generate":
		template, ok := dockerfileTemplates[analysis.Language]
		if !ok {
			return fmt.Errorf("unsupported language: %s", analysis.Language)
		}
		// override the default version in the template if the scanner found one (e.g. from .nvmrc or .python-version)
		if v := analysis.RuntimeVersion; v != "" {
			switch analysis.Language {
			case "node":
				template = strings.Replace(template, "node:20-alpine", "node:"+v+"-alpine", 1)
			case "python":
				template = strings.Replace(template, "python:3.11-slim", "python:"+v+"-slim", 1)
			}
		}
		dest := filepath.Join(repoPath, "Dockerfile")
		return os.WriteFile(dest, []byte(template), 0644)
	default:
		return fmt.Errorf("unsupported repo")
	}
}
