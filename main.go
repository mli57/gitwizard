package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/mli57/gitwizard/internal/fetch"
	"github.com/mli57/gitwizard/internal/scanner"
	"github.com/mli57/gitwizard/internal/generate"
	"github.com/mli57/gitwizard/internal/docker"
	"github.com/mli57/gitwizard/internal/state"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("usage: gitwizard <github-url>")
		os.Exit(1)
	}

	// FETCH: download and extract the repo zip locally
	result, err := fetch.FromURL(os.Args[1])
	if err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}

	fmt.Println("owner:", result.Owner)
	fmt.Println("repo: ", result.Repo)
	fmt.Println("path: ", result.Path)

	// SCAN: detect language, run method, and runtime version
	analysis, err := scanner.Scan(result.Path)
	if err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}

	fmt.Println("language: ", analysis.Language)
	fmt.Println("run method:", analysis.RunMethod)

	// each way of running a repo ends up with a set of host ports the app might be on 
	// which one it's actually on isn't known until we check them
	var hostPorts []int
	var containerID string
	var imageName string

	// 3 possibilities: compose.yml(auto build & run), standalone dockerfile, nothing related to docker
	if analysis.RunMethod == "compose" {
		// compose files leave blanks the repo expects you to fill in by hand
		notes, err := docker.PrepareEnv(result.Path)
		if err != nil {
			fmt.Println("error:", err)
			os.Exit(1)
		}
		if len(notes) > 0 {
			fmt.Println("filled in .env for you:")
			for _, note := range notes {
				fmt.Println(note)
			}
		}

		hostPorts, containerID, err = docker.RunCompose(result.Path)
		if err != nil {
			fmt.Println("error:", err)
			os.Exit(1)
		}
	} else if analysis.RunMethod == "generate" || analysis.RunMethod == "dockerfile" {
		if analysis.RunMethod == "generate" {
			// GENERATE: write a Dockerfile if the repo doesn't already have one
			err = generate.Dockerfile(result.Path, analysis)
			if err != nil {
				fmt.Println("error:", err)
				os.Exit(1)
			}

			fmt.Println("dockerfile ready")
		}

		// BUILD: build a Docker image from the Dockerfile
		imageName = strings.ToLower(result.Owner + "-" + result.Repo)
		err = docker.Build(result.Path, imageName)
		if err != nil {
			fmt.Println("error:", err)
			os.Exit(1)
		}

		// RUN: takes the docker image and runs it on a port
		fmt.Println("image built:", imageName)
		hostPorts, containerID, err = docker.Run(imageName)
		if err != nil {
			fmt.Println("error:", err)
			os.Exit(1)
		}
	}

	// READY: check the candidate ports until the app answers on one of them
	fmt.Println("checking ports:", hostPorts)
	hostPort, err := docker.WaitForApp(hostPorts)
	if err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}

	// STATE: record the installed app so it can be re-opened or uninstalled later
	appState, err := state.Load()
	if err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}
	key := strings.ToLower(result.Owner + "-" + result.Repo)
	appState.Apps[key] = state.AppEntry{
		ImageName:   imageName,
		ContainerID: containerID,
		Port:        hostPort,
		Path:        result.Path,
	}
	if err := state.Save(appState); err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}

	fmt.Println("app is running: http://localhost:" + fmt.Sprint(hostPort))
}
