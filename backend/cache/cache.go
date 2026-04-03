package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Manager struct {
	dir string
}

func NewManager(cacheDir string) *Manager {
	os.MkdirAll(cacheDir, 0755)
	return &Manager{dir: cacheDir}
}

func (m *Manager) HashContent(content string) string {
	h := sha256.Sum256([]byte(content))
	return hex.EncodeToString(h[:])
}

func (m *Manager) cacheFilePath(contentHash string) string {
	return filepath.Join(m.dir, contentHash+".json")
}

func (m *Manager) Check(content string, genre string) (*CacheResult, bool) {
	h := m.HashContent(content)
	path := m.cacheFilePath(h)

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, false
	}

	var entry struct {
		ContentHash       string `json:"contentHash"`
		Genre             string `json:"genre"`
		OutputPath        string `json:"outputPath"`
		TranslatedContent string `json:"translatedContent"`
		Timestamp         string `json:"timestamp"`
	}
	if err := json.Unmarshal(data, &entry); err != nil {
		return nil, false
	}

	if entry.Genre != genre {
		return nil, false
	}

	return &CacheResult{
		ContentHash:       h,
		TranslatedContent: entry.TranslatedContent,
		OutputPath:        entry.OutputPath,
		Timestamp:         entry.Timestamp,
	}, true
}

func (m *Manager) Save(contentHash string, genre string, translatedContent string, outputPath string, timestamp string) error {
	entry := map[string]string{
		"contentHash":       contentHash,
		"genre":             genre,
		"translatedContent": translatedContent,
		"outputPath":        outputPath,
		"timestamp":         timestamp,
	}

	data, err := json.MarshalIndent(entry, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(m.cacheFilePath(contentHash), data, 0644)
}

func (m *Manager) Dir() string {
	return m.dir
}

func (m *Manager) Clear() error {
	entries, err := os.ReadDir(m.dir)
	if err != nil {
		return err
	}

	for _, e := range entries {
		if e.Type().IsRegular() {
			if err := os.Remove(filepath.Join(m.dir, e.Name())); err != nil {
				return fmt.Errorf("failed to remove %s: %w", e.Name(), err)
			}
		}
	}

	return nil
}

func (m *Manager) GetCacheSize() (int64, error) {
	var total int64
	entries, err := os.ReadDir(m.dir)
	if err != nil {
		return 0, err
	}

	for _, e := range entries {
		if e.Type().IsRegular() {
			info, err := e.Info()
			if err == nil {
				total += info.Size()
			}
		}
	}

	return total, nil
}

type CacheResult struct {
	ContentHash       string
	TranslatedContent string
	OutputPath        string
	Timestamp         string
}
