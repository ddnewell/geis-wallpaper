// Package cache is a tiny on-disk store for fetched overlay data. It records
// when each entry was fetched so consumers can implement their own freshness
// policy: the fetcher decides whether to refetch (age vs TTL) and the renderer
// decides whether to draw or drop (age vs a hard "stale" limit). Writes are
// atomic (temp + rename) so the renderer never reads a half-written file.
package cache

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// Store is a flat directory of cached blobs plus sidecar ".meta" timestamps.
type Store struct {
	dir string
}

// New returns a Store rooted at dir (created on first write).
func New(dir string) *Store { return &Store{dir: dir} }

// Dir returns the store directory.
func (s *Store) Dir() string { return s.dir }

func (s *Store) path(name string) string     { return filepath.Join(s.dir, name) }
func (s *Store) metaPath(name string) string { return filepath.Join(s.dir, name+".meta") }

// Write atomically stores data under name and stamps the fetch time as now.
func (s *Store) Write(name string, data []byte, now time.Time) error {
	if err := os.MkdirAll(s.dir, 0o755); err != nil {
		return err
	}
	if err := atomicWrite(s.path(name), data); err != nil {
		return err
	}
	return atomicWrite(s.metaPath(name), []byte(strconv.FormatInt(now.Unix(), 10)))
}

// Read returns the cached bytes and the time they were fetched. ok is false if
// the entry is absent.
func (s *Store) Read(name string) (data []byte, fetchedAt time.Time, ok bool) {
	b, err := os.ReadFile(s.path(name))
	if err != nil {
		return nil, time.Time{}, false
	}
	return b, s.fetchedAt(name), true
}

// fetchedAt reads the sidecar timestamp, falling back to the file mtime.
func (s *Store) fetchedAt(name string) time.Time {
	if m, err := os.ReadFile(s.metaPath(name)); err == nil {
		if sec, err := strconv.ParseInt(strings.TrimSpace(string(m)), 10, 64); err == nil {
			return time.Unix(sec, 0)
		}
	}
	if fi, err := os.Stat(s.path(name)); err == nil {
		return fi.ModTime()
	}
	return time.Time{}
}

// Age returns how long ago name was fetched, and whether it exists.
func (s *Store) Age(name string, now time.Time) (time.Duration, bool) {
	if _, err := os.Stat(s.path(name)); err != nil {
		return 0, false
	}
	return now.Sub(s.fetchedAt(name)), true
}

func atomicWrite(path string, data []byte) error {
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}
