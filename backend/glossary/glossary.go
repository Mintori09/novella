package glossary

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"novella/backend/models"
)

type Manager struct {
	mu         sync.RWMutex
	glossaries map[string]*models.Glossary
	baseDir    string
}

func NewManager(baseDir string) *Manager {
	return &Manager{
		glossaries: make(map[string]*models.Glossary),
		baseDir:    baseDir,
	}
}

func (m *Manager) Get(novelID string) *models.Glossary {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if g, ok := m.glossaries[novelID]; ok {
		return g
	}
	return &models.Glossary{
		Entries: make(map[string]models.GlossaryEntry),
	}
}

func (m *Manager) Save(novelID string, glossary *models.Glossary) error {
	m.mu.Lock()
	m.glossaries[novelID] = glossary
	m.mu.Unlock()

	path := filepath.Join(m.baseDir, novelID+".json")
	if err := os.MkdirAll(m.baseDir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(glossary, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

func (m *Manager) Load(novelID string) (*models.Glossary, error) {
	path := filepath.Join(m.baseDir, novelID+".json")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			g := &models.Glossary{Entries: make(map[string]models.GlossaryEntry)}
			m.glossaries[novelID] = g
			return g, nil
		}
		return nil, err
	}

	var g models.Glossary
	if err := json.Unmarshal(data, &g); err != nil {
		return nil, err
	}

	if g.Entries == nil {
		g.Entries = make(map[string]models.GlossaryEntry)
	}

	m.mu.Lock()
	m.glossaries[novelID] = &g
	m.mu.Unlock()

	return &g, nil
}

func (m *Manager) AddEntry(novelID string, entry models.GlossaryEntry) error {
	g, err := m.Load(novelID)
	if err != nil {
		return err
	}

	g.Entries[entry.Original] = entry
	return m.Save(novelID, g)
}

func (m *Manager) RemoveEntry(novelID string, original string) error {
	g, err := m.Load(novelID)
	if err != nil {
		return err
	}

	delete(g.Entries, original)
	return m.Save(novelID, g)
}

func (m *Manager) ToPromptString(novelID string) string {
	g := m.Get(novelID)
	if len(g.Entries) == 0 {
		return ""
	}

	var parts []string
	parts = append(parts, "## Translation Glossary:")

	byCategory := make(map[string][]models.GlossaryEntry)
	for _, entry := range g.Entries {
		byCategory[entry.Category] = append(byCategory[entry.Category], entry)
	}

	for category, entries := range byCategory {
		parts = append(parts, "")
		parts = append(parts, "### "+strings.ToUpper(category[:1])+category[1:]+":")
		for _, entry := range entries {
			line := "- " + entry.Original + " → " + entry.Translated
			if entry.Note != "" {
				line += " (" + entry.Note + ")"
			}
			parts = append(parts, line)
		}
	}

	return strings.Join(parts, "\n")
}

func (m *Manager) ListNovels() ([]string, error) {
	entries, err := os.ReadDir(m.baseDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, err
	}

	var novels []string
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".json" {
			novels = append(novels, strings.TrimSuffix(entry.Name(), ".json"))
		}
	}

	return novels, nil
}

func (m *Manager) MergeEntries(novelID string, newEntries map[string]models.GlossaryEntry) error {
	g, err := m.Load(novelID)
	if err != nil {
		return err
	}

	for key, entry := range newEntries {
		if _, exists := g.Entries[key]; !exists {
			g.Entries[key] = entry
		}
	}

	return m.Save(novelID, g)
}

func ParseExtractedTerms(extractedText string) map[string]models.GlossaryEntry {
	entries := make(map[string]models.GlossaryEntry)

	lines := strings.Split(extractedText, "\n")
	currentCategory := "general"

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "###") || strings.HasPrefix(line, "##") {
			currentCategory = strings.TrimPrefix(line, "#")
			currentCategory = strings.TrimSpace(currentCategory)
			if currentCategory == "" {
				currentCategory = "general"
			}
			continue
		}

		if strings.HasPrefix(line, "-") || strings.HasPrefix(line, "*") {
			line = strings.TrimPrefix(line, "-")
			line = strings.TrimPrefix(line, "*")
			line = strings.TrimSpace(line)

			parts := strings.SplitN(line, "→", 2)
			if len(parts) != 2 {
				parts = strings.SplitN(line, "->", 2)
			}
			if len(parts) != 2 {
				continue
			}

			original := strings.TrimSpace(parts[0])
			rest := strings.TrimSpace(parts[1])

			translated := rest
			note := ""

			if idx := strings.Index(rest, "("); idx != -1 && strings.HasSuffix(rest, ")") {
				translated = strings.TrimSpace(rest[:idx])
				note = strings.TrimSuffix(rest[idx+1:], ")")
			}

			if original != "" && translated != "" {
				entries[original] = models.GlossaryEntry{
					Original:   original,
					Translated: translated,
					Category:   currentCategory,
					Note:       note,
				}
			}
		}
	}

	return entries
}
