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
		// only runs if a version file was found(.nvmrc, .python-version) and swaps the default runtime version in the template
		if v := analysis.RuntimeVersion; v != "" {
			switch analysis.Language {
			case "node":
				template = strings.Replace(template, "node:20-alpine", "node:"+v+"-alpine", 1)
			case "python":
				template = strings.Replace(template, "python:3.11-slim", "python:"+v+"-slim", 1)
			}
		}
		// always runs for python, swaps the default CMD with the right start command for the detected framework
		if analysis.Language == "python" {
			template = strings.Replace(template, `CMD ["python", "main.py"]`, pythonStartCmd(analysis.Framework), 1)
		}
		dest := filepath.Join(repoPath, "Dockerfile")
		return os.WriteFile(dest, []byte(template), 0644)
	default:
		return fmt.Errorf("unsupported repo")
	}
}

// takes the detected framework and returns the docker cmd line associated with it
func pythonStartCmd(framework string) string {
	switch framework {
	case "flask":
		return "ENV FLASK_APP=main.py\nCMD [\"flask\", \"run\", \"--host=0.0.0.0\", \"--port=5000\"]"
	case "fastapi":
		return "CMD [\"uvicorn\", \"main:app\", \"--host=0.0.0.0\", \"--port=5000\"]"
	case "django":
		return "CMD [\"python\", \"manage.py\", \"runserver\", \"0.0.0.0:8000\"]"
	default:
		return "CMD [\"python\", \"main.py\"]"
	}
}
