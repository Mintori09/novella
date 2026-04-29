package prompts

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"novella/backend/models"
)

var builtinGenres = map[string]string{
	"tien_hiep":   "Tiên hiệp",
	"ngon_tinh":   "Ngôn tình",
	"xuyen_khong": "Xuyên không",
	"phuong_tay":  "Phương Tây",
}

type Manager struct {
	embedFS   embed.FS
	customDir string
}

func NewManager(embedFS embed.FS, customDir string) *Manager {
	os.MkdirAll(customDir, 0o755)
	return &Manager{
		embedFS:   embedFS,
		customDir: customDir,
	}
}

func (m *Manager) ListGenres() []models.GenrePrompt {
	genres := make([]models.GenrePrompt, 0, len(builtinGenres))

	for id, name := range builtinGenres {
		content := m.getGenreContent(id)
		genres = append(genres, models.GenrePrompt{
			ID:      id,
			Name:    name,
			Content: content,
			Builtin: true,
		})
	}

	customFiles, _ := filepath.Glob(filepath.Join(m.customDir, "*.md"))
	for _, f := range customFiles {
		base := strings.TrimSuffix(filepath.Base(f), ".md")
		if base == "system_prompt" {
			continue
		}
		if _, isBuiltin := builtinGenres[base]; isBuiltin {
			continue
		}

		data, err := os.ReadFile(f)
		if err != nil {
			continue
		}

		name, content := parseCustomFile(string(data))
		if name == "" {
			name = formatLabel(base)
		}

		genres = append(genres, models.GenrePrompt{
			ID:      base,
			Name:    name,
			Content: content,
			Builtin: false,
		})
	}

	return genres
}

func (m *Manager) GetGenre(id string) (*models.GenrePrompt, error) {
	if name, ok := builtinGenres[id]; ok {
		content := m.getGenreContent(id)
		return &models.GenrePrompt{
			ID:      id,
			Name:    name,
			Content: content,
			Builtin: true,
		}, nil
	}

	path := filepath.Join(m.customDir, id+".md")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("genre '%s' not found", id)
	}

	name, content := parseCustomFile(string(data))
	if name == "" {
		name = formatLabel(id)
	}

	return &models.GenrePrompt{
		ID:      id,
		Name:    name,
		Content: content,
		Builtin: false,
	}, nil
}

func (m *Manager) AddGenre(id string, name string, content string) error {
	if _, ok := builtinGenres[id]; ok {
		return fmt.Errorf("genre '%s' already exists as a built-in genre", id)
	}

	path := filepath.Join(m.customDir, id+".md")
	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("genre '%s' already exists", id)
	}

	fileContent := buildCustomFile(name, content)
	return os.WriteFile(path, []byte(fileContent), 0o644)
}

func (m *Manager) UpdateGenre(id string, content string) error {
	if _, ok := builtinGenres[id]; ok {
		path := filepath.Join(m.customDir, id+".md")
		existing, err := m.GetGenre(id)
		if err != nil {
			return err
		}
		return os.WriteFile(path, []byte(buildCustomFile(existing.Name, content)), 0o644)
	}

	path := filepath.Join(m.customDir, id+".md")
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("genre '%s' not found", id)
	}

	name, _ := parseCustomFile(string(data))
	if name == "" {
		name = formatLabel(id)
	}

	return os.WriteFile(path, []byte(buildCustomFile(name, content)), 0o644)
}

func (m *Manager) RemoveGenre(id string) error {
	if _, ok := builtinGenres[id]; ok {
		return fmt.Errorf("cannot remove built-in genre '%s'", id)
	}

	path := filepath.Join(m.customDir, id+".md")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("genre '%s' not found", id)
	}

	return os.Remove(path)
}

func (m *Manager) GetSystemPrompt() string {
	if _, ok := builtinGenres["tien_hiep"]; ok {
		data, err := m.embedFS.ReadFile("system_prompt.md")
		if err == nil {
			return string(data)
		}
	}

	path := filepath.Join(m.customDir, "system_prompt.md")
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return string(data)
}

func (m *Manager) GetGenrePrompt(id string) (string, error) {
	if _, ok := builtinGenres[id]; ok {
		data, err := m.embedFS.ReadFile(id + ".md")
		if err != nil {
			return "", err
		}
		return string(data), nil
	}

	path := filepath.Join(m.customDir, id+".md")
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	_, content := parseCustomFile(string(data))
	return content, nil
}

func (m *Manager) getGenreContent(id string) string {
	data, err := m.embedFS.ReadFile(id + ".md")
	if err != nil {
		return ""
	}
	return string(data)
}

func parseCustomFile(content string) (name string, body string) {
	content = strings.TrimSpace(content)
	if strings.HasPrefix(content, "---") {
		parts := strings.SplitN(content[3:], "---", 2)
		if len(parts) == 2 {
			frontmatter := strings.TrimSpace(parts[0])
			body = strings.TrimSpace(parts[1])
			for _, line := range strings.Split(frontmatter, "\n") {
				line = strings.TrimSpace(line)
				if strings.HasPrefix(line, "name:") {
					name = strings.TrimSpace(strings.TrimPrefix(line, "name:"))
				}
			}
			return name, body
		}
	}
	return "", content
}

func buildCustomFile(name string, content string) string {
	return fmt.Sprintf("---\nname: %s\n---\n\n%s", name, content)
}

func formatLabel(id string) string {
	parts := strings.Split(id, "_")
	for i, p := range parts {
		if len(p) > 0 {
			parts[i] = strings.ToUpper(p[:1]) + p[1:]
		}
	}
	return strings.Join(parts, " ")
}
