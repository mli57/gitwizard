package main

import (
	"fmt"
	"os"

	"github.com/mli57/gitwizard/internal/fetch"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("usage: gitwizard <github-url>")
		os.Exit(1)
	}

	result, err := fetch.FromURL(os.Args[1])
	if err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}

	fmt.Println("owner:", result.Owner)
	fmt.Println("repo: ", result.Repo)
	fmt.Println("path: ", result.Path)
}
