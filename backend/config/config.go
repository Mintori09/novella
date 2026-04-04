package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"

	"novella/backend/models"
)

type Manager struct {
	mu     sync.RWMutex
	config *models.AppConfig
	path   string
}

func NewManager(configPath string) (*Manager, error) {
	m := &Manager{
		path: configPath,
	}

	if err := m.Load(); err != nil {
		m.config = defaultConfig()
		if err := m.Save(); err != nil {
			return nil, err
		}
	}

	return m, nil
}

func defaultConfig() *models.AppConfig {
	homeDir, _ := os.UserHomeDir()
	return &models.AppConfig{
		Providers: map[string]models.ProviderConfig{
			"kilo": {
				APIKey:       "",
				Model:        "kilo-auto/free",
				Models:       []string{"kilo-auto/free"},
				DefaultModel: "kilo-auto/free",
				Enabled:      true,
			},
			"openrouter": {
				APIKey:       "",
				Model:        "google/gemini-2.0-flash-exp:free",
				Models:       []string{"google/gemini-2.0-flash-exp:free"},
				DefaultModel: "google/gemini-2.0-flash-exp:free",
				Enabled:      false,
			},
			"openai": {
				APIKey:       "",
				Model:        "gpt-4o-mini",
				Models:       []string{"gpt-4o-mini"},
				DefaultModel: "gpt-4o-mini",
				Enabled:      false,
			},
			"groq": {
				APIKey:       "",
				Model:        "llama-3.3-70b-versatile",
				Models:       []string{"llama-3.3-70b-versatile"},
				DefaultModel: "llama-3.3-70b-versatile",
				Enabled:      false,
			},
			"gemini": {
				APIKey:       "",
				Model:        "gemini-2.0-flash",
				Models:       []string{"gemini-2.0-flash"},
				DefaultModel: "gemini-2.0-flash",
				Enabled:      false,
			},
		},
		ActiveProvider:         "kilo",
		WorkerCount:            4,
		ChunkSize:              3500,
		EnableReview:           true,
		OutputDir:              filepath.Join(homeDir, "novel-translate", "output"),
		OutputFormat:           "txt",
		FallbackOnFailure:      true,
		Theme:                  "dark",
		CustomPromptDir:        "",
		CustomPromptDirEnabled: false,
		MaxRetries:             2,
	}
}

func (m *Manager) Load() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	data, err := os.ReadFile(m.path)
	if err != nil {
		return err
	}

	var cfg models.AppConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return err
	}

	m.config = &cfg
	return nil
}

func (m *Manager) Save() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	dir := filepath.Dir(m.path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(m.config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(m.path, data, 0644)
}

func (m *Manager) Get() *models.AppConfig {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.config
}

func (m *Manager) Update(fn func(*models.AppConfig)) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	fn(m.config)

	data, err := json.MarshalIndent(m.config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(m.path, data, 0644)
}

func (m *Manager) GetProvider(id string) (models.ProviderConfig, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	p, ok := m.config.Providers[id]
	return p, ok
}

func (m *Manager) GetActiveProvider() (string, models.ProviderConfig, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	p, ok := m.config.Providers[m.config.ActiveProvider]
	return m.config.ActiveProvider, p, ok
}
