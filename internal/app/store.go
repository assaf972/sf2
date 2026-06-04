package app

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Store persists scenes (performance patches) as JSON files in a directory.
type Store struct {
	dir string
}

// NewStore returns a scene store rooted at dir, creating it if needed.
func NewStore(dir string) (*Store, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	return &Store{dir: dir}, nil
}

// DefaultDir returns the per-user scenes directory.
func DefaultDir() string {
	if cfg, err := os.UserConfigDir(); err == nil {
		return filepath.Join(cfg, "gigsynth", "scenes")
	}
	return "scenes"
}

func (s *Store) path(name string) string {
	safe := strings.ReplaceAll(name, string(os.PathSeparator), "_")
	return filepath.Join(s.dir, safe+".json")
}

// List returns the names of saved scenes.
func (s *Store) List() ([]string, error) {
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		return nil, err
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		names = append(names, strings.TrimSuffix(e.Name(), ".json"))
	}
	sort.Strings(names)
	return names, nil
}

// Save writes a scene under its Name.
func (s *Store) Save(scene Scene) error {
	if strings.TrimSpace(scene.Name) == "" {
		return fmt.Errorf("scene name is empty")
	}
	data, err := json.MarshalIndent(scene, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path(scene.Name), data, 0o644)
}

// Load reads a scene by name.
func (s *Store) Load(name string) (Scene, error) {
	var scene Scene
	data, err := os.ReadFile(s.path(name))
	if err != nil {
		return scene, err
	}
	if err := json.Unmarshal(data, &scene); err != nil {
		return scene, err
	}
	if len(scene.Layers) == 0 {
		return scene, fmt.Errorf("scene %q has no layers", name)
	}
	return scene, nil
}

// Delete removes a saved scene.
func (s *Store) Delete(name string) error {
	return os.Remove(s.path(name))
}
