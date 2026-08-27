package infrastructure

import (
	"os"
	"path/filepath"
	"testing"

	"example.com/reproducible-build-farm/internal/domain"
)

func TestSaveCleansTempOnError(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "snap.json")
	if err := os.Mkdir(path, 0755); err != nil {
		t.Fatal(err)
	}
	fs := NewFileSnapshot(path)
	err := fs.Save(Snapshot{Projects: map[string]domain.Project{}})
	if err == nil {
		t.Fatal("expected save to fail when target is a directory")
	}
	if _, statErr := os.Stat(path + ".tmp"); !os.IsNotExist(statErr) {
		t.Fatal("temporary file leaked after failed save")
	}
}
