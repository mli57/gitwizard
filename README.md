# GitWizard
(Ongoing project, things may change as I continue to work on it)

A Windows desktop app that lets anyone install and run any GitHub repo with one click: this means no terminal, no installs, no config. Created for those of who are not technical but still want to use the tools offered on GitHub!

## What it is

GitWizard is a local `.exe` that takes a GitHub URL and gets the app running in your browser. You paste a github repository link. It handles everything else.

Underneath, it's a Go backend that serves a local React UI. Docker is the execution sandbox. Since we are running unverified code from the internet. Also because Docker can reproduce any environment without the user installing Node, Python, Go, or anything else.

## What it does

1. **Paste a URL:** for example `https://github.com/some-user/some-app`
2. **Fetches the repo:** downloads a zip directly from GitHub, no git required
3. **Scans the code:** detects the language, framework, and port from local files
4. **Decides a strategy:** uses an existing `docker-compose.yml` or `Dockerfile` if present; otherwise generates one from a template
5. **Builds a Docker image:** streams the build log to the UI in real time
6. **Runs the container:** maps the right port, injects env vars, enforces resource limits
7. **Opens the browser:** waits until the app actually responds, then opens it

*If anything fails at any step, a log of the error is displayed with an explanation.

## Why it's useful

Running a GitHub repo from scratch normally requires: knowing what runtime to install, which version, what commands to run, what port to visit, and how to debug it when something goes wrong. For a non-technical user, that's a lot to learn.

GitWizard removes that knowledge gap. The target user is someone who found a cool open-source app and just wants to use it, not a developer setting up a project. The point of this project is to make GitHub a more accessible to everyone, not just a place for developers to dump their projects.

## What's coming up for v1

- Public GitHub repos, default branch only
- Node, Python, and Go web apps that are self-contained
- Repos that ship their own `docker-compose.yml` (including those with a database sidecar)
- API key prompting if the repo has a `.env.example`
- Windows only

Out of scope: private repos, remote databases, monorepos, non-web apps.

## Current stage

**`internal/fetch` (complete.)**

The fetcher (`internal/fetch/fetch.go`) is written and tested end to end:
- Parses a GitHub URL into owner + repo
- Downloads the repo zip from the GitHub API
- Extracts it to `%LOCALAPPDATA%\GitWizard\apps\<owner>-<repo>\`, stripping GitHub's wrapper folder

**Next: `internal/scanner`** 
walks the extracted repo and produces a `RepoAnalysis` describing the stack (language, framework, port, whether Docker assets exist, etc.). That output drives everything downstream.

## Architecture

| Package | What it does | Status |
|---|---|---|
| `fetch` | Download & extract repo zip | Done |
| `scanner` | Detect stack from local files | Coming up next |
| `generate` | Render Dockerfile from template | Pending |
| `docker` | Build, run, stream logs, readiness | Pending |
| `state` | Track installed apps (JSON) | Pending |
| `api` | HTTP API for the React frontend | Pending |

Packaged as a single `.exe` using `go build` with `go:embed` for the React assets.

## Requirements

- Windows 10/11
- [Docker Desktop](https://www.docker.com/products/docker-desktop/) (GitWizard will prompt if it's not installed)
