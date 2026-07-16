package importer

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	ocsync "github.com/dagimg-dot/oc-sync/internal/sync"
	"gopkg.in/yaml.v3"
)

type peerState map[string]fileInfo // session_id → info

type fileInfo struct {
	Size      int64 `yaml:"size"`
	MtimeNano int64 `yaml:"mtime_nano"`
}

// Tracker records which peer files have been imported by content fingerprint
// (file size + mtime). If all three match, the file hasn't changed since import.
type Tracker struct {
	mu       sync.Mutex
	path     string
	Imported map[string]peerState `yaml:"imported"` // hostname → session_id → info
}

func NewTracker(configDir string) (*Tracker, error) {
	t := &Tracker{
		path:     filepath.Join(configDir, "sync-state.yaml"),
		Imported: make(map[string]peerState),
	}

	f, err := os.Open(t.path)
	if err != nil {
		if os.IsNotExist(err) {
			return t, nil
		}
		return nil, fmt.Errorf("open sync state: %w", err)
	}
	defer f.Close()

	if err := yaml.NewDecoder(f).Decode(t); err != nil {
		return nil, fmt.Errorf("decode sync state: %w", err)
	}
	if t.Imported == nil {
		t.Imported = make(map[string]peerState)
	}
	return t, nil
}

func (t *Tracker) IsImported(peerHostname, path string) bool {
	id := ocsync.SessionID(path)
	info, err := os.Stat(path)
	if err != nil {
		return false
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	hostState, ok := t.Imported[peerHostname]
	if !ok {
		return false
	}
	prev, ok := hostState[id]
	if !ok {
		return false
	}
	return prev.Size == info.Size() && prev.MtimeNano == info.ModTime().UnixNano()
}

func (t *Tracker) MarkImported(peerHostname, path string) {
	id := ocsync.SessionID(path)
	info, err := os.Stat(path)
	if err != nil {
		return
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	hostState, ok := t.Imported[peerHostname]
	if !ok {
		hostState = make(peerState)
		t.Imported[peerHostname] = hostState
	}
	hostState[id] = fileInfo{
		Size:      info.Size(),
		MtimeNano: info.ModTime().UnixNano(),
	}
}

func (t *Tracker) Save() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	dir := filepath.Dir(t.path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("mkdir state dir: %w", err)
	}

	f, err := os.Create(t.path)
	if err != nil {
		return fmt.Errorf("create sync state: %w", err)
	}
	defer f.Close()

	enc := yaml.NewEncoder(f)
	enc.SetIndent(2)
	if err := enc.Encode(t); err != nil {
		return fmt.Errorf("encode sync state: %w", err)
	}
	return enc.Close()
}
