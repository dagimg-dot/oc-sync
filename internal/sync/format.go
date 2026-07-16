package sync

import (
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const (
	ExtNew = ".json.gz"
	ExtOld = ".json"
)

func SessionFileName(id string) string {
	return id + ExtNew
}

func SessionID(name string) string {
	base := filepath.Base(name)
	if strings.HasSuffix(base, ExtNew) {
		return strings.TrimSuffix(base, ExtNew)
	}
	return strings.TrimSuffix(base, ExtOld)
}

func IsSessionFile(name string) bool {
	return strings.HasSuffix(name, ExtNew) || strings.HasSuffix(name, ExtOld)
}

type gzipReadCloser struct {
	gr *gzip.Reader
	f  *os.File
}

func (g *gzipReadCloser) Read(p []byte) (int, error) { return g.gr.Read(p) }
func (g *gzipReadCloser) Close() error               { g.gr.Close(); return g.f.Close() }

func OpenSessionFile(path string) (io.ReadCloser, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	if strings.HasSuffix(path, ExtNew) {
		gr, err := gzip.NewReader(f)
		if err != nil {
			f.Close()
			return nil, err
		}
		return &gzipReadCloser{gr, f}, nil
	}
	return f, nil
}
