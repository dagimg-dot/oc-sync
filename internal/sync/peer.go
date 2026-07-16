package sync

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"strings"
)

var bg = context.Background()

type PeerFile struct {
	Machine string
	Path    string
}

func PeerFiles(syncDir, ownHostname string) ([]PeerFile, error) {
	entries, err := os.ReadDir(syncDir)
	if err != nil {
		return nil, err
	}
	var files []PeerFile
	for _, e := range entries {
		if !e.IsDir() || e.Name() == ownHostname {
			continue
		}
		machineDir := filepath.Join(syncDir, e.Name())
		dirEntries, err := os.ReadDir(machineDir)
		if err != nil {
			continue
		}
		for _, f := range dirEntries {
			if f.IsDir() || !strings.HasSuffix(f.Name(), ".json") {
				continue
			}
			files = append(files, PeerFile{
				Machine: e.Name(),
				Path:    filepath.Join(machineDir, f.Name()),
			})
		}
	}
	return files, nil
}

func PendingExports(db *sql.DB, myDir string) (int, error) {
	rows, err := db.QueryContext(bg, "SELECT id FROM session")
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	var count int
	for rows.Next() {
		var id string
		rows.Scan(&id)
		p := filepath.Join(myDir, id+".json")
		if _, err := os.Stat(p); os.IsNotExist(err) {
			count++
		}
	}
	return count, rows.Err()
}

func PendingImports(syncDir, ownHostname string) (int, error) {
	files, err := PeerFiles(syncDir, ownHostname)
	if err != nil {
		return 0, nil
	}
	return len(files), nil
}
