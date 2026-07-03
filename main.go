package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/mli57/gitwizard/src/fetch"
	"github.com/mli57/gitwizard/src/scanner"
	"github.com/mli57/gitwizard/src/generate"
	"github.com/mli57/gitwizard/src/docker"
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

	// GENERATE: write a Dockerfile if the repo doesn't already have one
	err = generate.Dockerfile(result.Path, analysis)
	if err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}

	fmt.Println("dockerfile ready")

	// BUILD: build a Docker image from the Dockerfile
	imageName := strings.ToLower(result.Owner + "-" + result.Repo)
	err = docker.Build(result.Path, imageName)
	if err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}

	// RUN: takes the docker image and runs it on a port
	fmt.Println("image built:", imageName)
	hostPort, err := docker.Run(imageName, 5000)
	if err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}
	fmt.Println("Host port: ", hostPort)

	// READY: polling app until it responds
	err = docker.WaitForApp(hostPort)
	if err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}
}
