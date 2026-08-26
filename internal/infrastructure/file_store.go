package infrastructure

import (
	"encoding/json"
	"example.com/reproducible-build-farm/internal/domain"
	"os"
	"path/filepath"
	"sync"
)

type Snapshot struct {
	Projects    map[string]domain.Project         `json:"projects"`
	Definitions map[string]domain.BuildDefinition `json:"definitions"`
	Executions  map[string]domain.Execution       `json:"executions"`
}
type FileSnapshot struct {
	mu   sync.Mutex
	path string
}

func NewFileSnapshot(path string) *FileSnapshot { return &FileSnapshot{path: path} }
func (f *FileSnapshot) Save(s Snapshot) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if err := os.MkdirAll(filepath.Dir(f.path), 0750); err != nil {
		return err
	}
	tmp := f.path + ".tmp"
	b, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	if err = os.WriteFile(tmp, b, 0600); err != nil {
		return err
	}
	return os.Rename(tmp, f.path)
}
func (f *FileSnapshot) Load() (Snapshot, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	b, err := os.ReadFile(f.path)
	if err != nil {
		return Snapshot{}, err
	}
	var s Snapshot
	err = json.Unmarshal(b, &s)
	return s, err
}
