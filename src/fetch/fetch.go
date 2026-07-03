/*
Pulls repositories from GitHub and installs onto local machine
*/
package fetch

import (
	"archive/zip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type Result struct {
	Owner string
	Repo string
	Path string
}

// parseGitHubURL pulls owner and repo out of a URL like https://github.com/owner/repo
func parseGitHubURL(rawURL string) (owner, repo string, err error){
	rawURL = strings.TrimRight(rawURL, "/")
	rawURL = strings.TrimSuffix(rawURL, ".git")

	// splitting "https://github.com/owner/repo" by "/" gives: ["https:", "", "github.com", "owner", "repo"]
	urlParts := strings.Split(rawURL, "/")
	if len(urlParts) != 5 || urlParts[2] != "github.com"{
		return "", "", fmt.Errorf("Url is incorrect. Expected https://github.com/owner/repo, got: %s", rawURL)
	}
	return urlParts[3], urlParts[4], nil
}

// Creates & returns a path to house data downloaded from repo
func installDir(owner, repo string) (string, error){
	appData := os.Getenv("LOCALAPPDATA")
	if appData == "" {
		return "", fmt.Errorf("LOCALAPPDATA not set")
	}

	// builds %LOCALAPPDATA%\GitWizard\apps\<owner>-<repo> w/ os.MkdirAll
	dir := filepath.Join(appData, "GitWizard", "apps", owner+"-"+repo)
	err := os.MkdirAll(dir, 0755)
	if err != nil{
		return "", fmt.Errorf("could not create install dir: %w", err)
	}
	return dir, nil
}

// fetches the repo zip from GitHub and writes it to destPath
func downloadRepo(owner, repo, destPath string) error{
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/zipball", owner, repo)

	// make the GET request, github redirects to the actual zip
	response, err := http.Get(url)
	if err != nil{
		return fmt.Errorf("Download failed: %w", err)
	}
	defer response.Body.Close() // defer early

	// check status
	if response.StatusCode == 404 {
		return fmt.Errorf("repo not found: github.com/%s/%s", owner, repo)
	}
	if response.StatusCode != 200 {
		return fmt.Errorf("unexpected status %d", response.StatusCode)
	}

	// create the file on disk that we'll write the zip into
	destFile, err := os.Create(destPath)
	if err != nil{
		return fmt.Errorf("could not create file: %w", err)
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, response.Body) // copies .zip to destPath
	return err
}

// unzips zipPath into destDir, while also stripping GitHub's folder prefix
func extract(zipPath, destDir string) error{
	reader, err := zip.OpenReader(zipPath)
	if err != nil{
		return fmt.Errorf("could not open zip: %w", err)
	}
	defer reader.Close()

	// gets name of the first file, cut off the trailing
	var prefix string
	if len(reader.File) > 0{
		firstEntry := reader.File[0].Name
		parts := strings.SplitN(firstEntry, "/", 2)
		prefix = parts[0] + "/"
	}

	// for every entry(file, directory), move to a new repo w/o the trailing numbers
	for _, entry := range reader.File{
		strippedPath := strings.TrimPrefix(entry.Name, prefix)
		if strippedPath == "" {
			continue // skip the root folder entry
		}

		destPath := filepath.Join(destDir, strippedPath)

		// If entry is a folder, create it and continue
		if entry.FileInfo().IsDir() {
			os.MkdirAll(destPath, entry.Mode())
			continue
		}

		os.MkdirAll(filepath.Dir(destPath), 0755) // ensure parent dirs exist

		err := writeFile(entry, destPath)
		if err != nil {
			return err
		}
	}
	return nil
}

// extracts a single file from the zip to destPath
func writeFile(entry *zip.File, destPath string) error{
	zipEntry, err := entry.Open() // opens file for reading
	if err != nil{
		return err
	}
	defer zipEntry.Close()

	destination, err := os.Create(destPath) // create destination file
	if err != nil{
		return err
	}
	defer destination.Close()

	_, err = io.Copy(destination, zipEntry) // copy file to new dir
	return err
}

func FromURL(rawURL string) (Result, error) {
	owner, repo, err := parseGitHubURL(rawURL)
	if err != nil{
		return Result{}, err
	}

	directory, err := installDir(owner, repo)
	if err != nil{
		return Result{}, err
	}

	zipPath := filepath.Join(directory, "repo.zip")
	err = downloadRepo(owner, repo, zipPath)
	if err != nil{
		return Result{}, err
	}

	err = extract(zipPath, directory)
	if err != nil{
		return Result{}, err
	}

	return Result{Owner: owner, Repo: repo, Path: directory}, nil
}