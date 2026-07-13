/*
Tracks installed apps in %LOCALAPPDATA%\GitWizard\state.json so they can be
re-opened, listed, and uninstalled instead of leaking containers and images
*/
package state

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type AppEntry struct {
	ImageName   string
	ContainerID string
	Port        int
	Path        string
}

type State struct {
	Apps map[string]AppEntry // keyed by "<owner>-<repo>"
}

// returns the path to the one global state file: %LOCALAPPDATA%\GitWizard\state.json
func statePath() (string, error) {
	appData := os.Getenv("LOCALAPPDATA")
	if appData == "" {
		return "", fmt.Errorf("LOCALAPPDATA not set")
	}
	return filepath.Join(appData, "GitWizard", "state.json"), nil
}

// reads the state file; a missing file is not an error, it just means no apps yet
func Load() (State, error) {
	path, err := statePath()
	if err != nil {
		return State{}, err
	}

	b, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return State{Apps: map[string]AppEntry{}}, nil
	}
	if err != nil {
		return State{}, fmt.Errorf("could not read state file: %w", err)
	}

	var s State
	if err := json.Unmarshal(b, &s); err != nil {
		return State{}, fmt.Errorf("state file is corrupted: %w", err)
	}
	if s.Apps == nil {
		s.Apps = map[string]AppEntry{}
	}
	return s, nil
}

// writes the whole state back to disk
func Save(s State) error {
	path, err := statePath()
	if err != nil {
		return err
	}

	// make sure the GitWizard folder exists (first run may save before any fetch)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	b, err := json.MarshalIndent(s, "", "  ") // indented so the file is human-readable
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0644)
}
