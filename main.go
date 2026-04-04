package main

import (
	"context"
	"embed"
	"fmt"
	"os"
	"path/filepath"

	"novella/backend/cache"
	"novella/backend/config"
	"novella/backend/engine"
	"novella/backend/glossary"
	"novella/backend/models"
	promptsmgr "novella/backend/prompts"
	"novella/prompts"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/linux"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

//go:embed all:frontend/dist
var assets embed.FS

type App struct {
	ctx       context.Context
	cfg       *config.Manager
	engine    *engine.Engine
	glossary  *glossary.Manager
	cacheMgr  *cache.Manager
	promptMgr *promptsmgr.Manager
}

func NewApp() *App {
	homeDir, _ := os.UserHomeDir()
	configDir := filepath.Join(homeDir, ".config", "novella")
	os.MkdirAll(configDir, 0755)

	cfgPath := filepath.Join(configDir, "config.json")
	cfg, _ := config.NewManager(cfgPath)

	glossaryDir := filepath.Join(configDir, "glossaries")
	glossaryMgr := glossary.NewManager(glossaryDir)

	cacheDir := filepath.Join(configDir, "cache")
	cacheMgr := cache.NewManager(cacheDir)

	promptDir := filepath.Join(configDir, "genres")
	promptMgr := promptsmgr.NewManager(prompts.FS, promptDir)

	eng := engine.NewEngine(cfg, glossaryMgr, cacheMgr, promptMgr)

	return &App{
		cfg:       cfg,
		engine:    eng,
		glossary:  glossaryMgr,
		cacheMgr:  cacheMgr,
		promptMgr: promptMgr,
	}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

func (a *App) GetConfig() *models.AppConfig {
	return a.cfg.Get()
}

func (a *App) UpdateConfig(updates map[string]interface{}) error {
	return a.cfg.Update(func(cfg *models.AppConfig) {
		if v, ok := updates["activeProvider"].(string); ok {
			cfg.ActiveProvider = v
		}
		if v, ok := updates["workerCount"].(float64); ok {
			cfg.WorkerCount = int(v)
		}
		if v, ok := updates["chunkSize"].(float64); ok {
			cfg.ChunkSize = int(v)
		}
		if v, ok := updates["enableReview"].(bool); ok {
			cfg.EnableReview = v
		}
		if v, ok := updates["outputDir"].(string); ok {
			cfg.OutputDir = v
		}
		if v, ok := updates["outputFormat"].(string); ok {
			cfg.OutputFormat = v
		}
		if v, ok := updates["maxRetries"].(float64); ok {
			cfg.MaxRetries = int(v)
		}
		if v, ok := updates["theme"].(string); ok {
			cfg.Theme = v
		}
		if v, ok := updates["customPromptDir"].(string); ok {
			cfg.CustomPromptDir = v
		}
		if v, ok := updates["customPromptDirEnabled"].(bool); ok {
			cfg.CustomPromptDirEnabled = v
		}
		if providers, ok := updates["providers"].(map[string]interface{}); ok {
			for id, p := range providers {
				if pm, ok := p.(map[string]interface{}); ok {
					prov := cfg.Providers[id]
					if apiKey, ok := pm["apiKey"].(string); ok {
						prov.APIKey = apiKey
					}
					if model, ok := pm["model"].(string); ok {
						prov.Model = model
					}
					if models, ok := pm["models"].([]interface{}); ok {
						prov.Models = make([]string, len(models))
						for i, m := range models {
							if s, ok := m.(string); ok {
								prov.Models[i] = s
							}
						}
					}
					if defaultModel, ok := pm["defaultModel"].(string); ok {
						prov.DefaultModel = defaultModel
					}
					if enabled, ok := pm["enabled"].(bool); ok {
						prov.Enabled = enabled
					}
					cfg.Providers[id] = prov
				}
			}
		}
	})
}

func (a *App) GetProviders() map[string]models.APIProvider {
	return map[string]models.APIProvider{
		"kilo": {
			ID:           "kilo",
			Name:         "Kilo AI",
			BaseURL:      "https://api.kilo.ai/api/gateway",
			DefaultModel: "kilo-auto/free",
			APIKeyEnv:    "KILO_API_KEY",
		},
		"openrouter": {
			ID:           "openrouter",
			Name:         "OpenRouter",
			BaseURL:      "https://openrouter.ai/api/v1",
			DefaultModel: "google/gemini-2.0-flash-exp:free",
			APIKeyEnv:    "OPENROUTER_API_KEY",
		},
		"openai": {
			ID:           "openai",
			Name:         "OpenAI",
			BaseURL:      "https://api.openai.com/v1",
			DefaultModel: "gpt-4o-mini",
			APIKeyEnv:    "OPENAI_API_KEY",
		},
		"groq": {
			ID:           "groq",
			Name:         "Groq",
			BaseURL:      "https://api.groq.com/openai/v1",
			DefaultModel: "llama-3.3-70b-versatile",
			APIKeyEnv:    "GROQ_API_KEY",
		},
		"gemini": {
			ID:           "gemini",
			Name:         "Google Gemini",
			BaseURL:      "https://generativelanguage.googleapis.com/v1beta/openai",
			DefaultModel: "gemini-2.0-flash",
			APIKeyEnv:    "GEMINI_API_KEY",
		},
	}
}

func (a *App) TestConnection(providerID string, apiKey string, model string) *models.TestConnectionResult {
	result, _ := a.engine.TestProviderConnection(a.ctx, providerID, apiKey, model)
	return result
}

func (a *App) TranslateFile(inputPath string, genre string) {
	go a.engine.TranslateFile(a.ctx, inputPath, genre)
}

func (a *App) TranslateFiles(paths []string, genre string) {
	go a.engine.TranslateFiles(a.ctx, paths, genre)
}

func (a *App) CancelTask(taskID string) {
	a.engine.CancelTask(taskID)
}

func (a *App) StopAll() {
	a.engine.StopAll()
}

func (a *App) GetTasks() []*models.TranslationTask {
	return a.engine.GetTasks()
}

func (a *App) DirectoryExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

func (a *App) SelectDirectory() (string, error) {
	dir, err := runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Select Directory",
	})
	return dir, err
}

func (a *App) SelectFile(filter string) (string, error) {
	return runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Select File",
		Filters: []runtime.FileFilter{
			{DisplayName: filter, Pattern: "*." + filter},
		},
	})
}

func (a *App) SelectFiles() ([]string, error) {
	return runtime.OpenMultipleFilesDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Select Files",
		Filters: []runtime.FileFilter{
			{DisplayName: "Text Files", Pattern: "*.txt;*.md"},
		},
	})
}

func (a *App) ListDirectory(dirPath string) ([]models.FileInfo, error) {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, err
	}

	var files []models.FileInfo
	for _, entry := range entries {
		info, _ := entry.Info()
		files = append(files, models.FileInfo{
			Name:  entry.Name(),
			Path:  filepath.Join(dirPath, entry.Name()),
			Size:  info.Size(),
			IsDir: entry.IsDir(),
		})
	}
	return files, nil
}

func (a *App) GetGlossary(novelID string) *models.Glossary {
	g, _ := a.glossary.Load(novelID)
	return g
}

func (a *App) SaveGlossary(novelID string, glossary *models.Glossary) error {
	return a.glossary.Save(novelID, glossary)
}

func (a *App) AddGlossaryEntry(novelID string, entry models.GlossaryEntry) error {
	return a.glossary.AddEntry(novelID, entry)
}

func (a *App) RemoveGlossaryEntry(novelID string, original string) error {
	return a.glossary.RemoveEntry(novelID, original)
}

func (a *App) ListGlossaries() ([]string, error) {
	return a.glossary.ListNovels()
}

func (a *App) GetGenreList() []models.GenrePrompt {
	return a.promptMgr.ListGenres()
}

func (a *App) AddGenrePrompt(id string, name string, content string) error {
	return a.promptMgr.AddGenre(id, name, content)
}

func (a *App) UpdateGenrePrompt(id string, content string) error {
	return a.promptMgr.UpdateGenre(id, content)
}

func (a *App) RemoveGenrePrompt(id string) error {
	return a.promptMgr.RemoveGenre(id)
}

func (a *App) GetGenrePrompt(id string) (*models.GenrePrompt, error) {
	return a.promptMgr.GetGenre(id)
}

func (a *App) ClearCache() error {
	return a.cacheMgr.Clear()
}

func (a *App) GetCacheSize() (int64, error) {
	return a.cacheMgr.GetCacheSize()
}

type CacheInfo struct {
	Size      int64  `json:"size"`
	SizeHuman string `json:"sizeHuman"`
}

func (a *App) GetCacheInfo() (*CacheInfo, error) {
	size, err := a.cacheMgr.GetCacheSize()
	if err != nil {
		return nil, err
	}
	return &CacheInfo{
		Size:      size,
		SizeHuman: formatBytes(size),
	}, nil
}

func formatBytes(b int64) string {
	if b < 1024 {
		return fmt.Sprintf("%d B", b)
	}
	if b < 1024*1024 {
		return fmt.Sprintf("%.1f KB", float64(b)/1024)
	}
	return fmt.Sprintf("%.1f MB", float64(b)/(1024*1024))
}

type DroppedFile struct {
	Name    string `json:"name"`
	Content string `json:"content"`
}

func (a *App) ProcessDroppedFiles(droppedFiles []DroppedFile) (string, error) {
	tempDir := filepath.Join(a.cacheMgr.Dir(), "drop-temp")
	os.RemoveAll(tempDir)
	os.MkdirAll(tempDir, 0755)

	for _, f := range droppedFiles {
		if f.Name == "" {
			continue
		}
		safeName := filepath.Base(f.Name)
		path := filepath.Join(tempDir, safeName)
		if err := os.WriteFile(path, []byte(f.Content), 0644); err != nil {
			return "", err
		}
	}

	return tempDir, nil
}

func main() {
	app := NewApp()

	err := wails.Run(&options.App{
		Title:            "novella",
		Width:            1280,
		Height:           800,
		MinWidth:         900,
		MinHeight:        600,
		Frameless:        true,
		DisableResize:    false,
		BackgroundColour: &options.RGBA{R: 0, G: 0, B: 0, A: 1},
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		OnStartup: app.startup,
		Bind: []interface{}{
			app,
		},
		Linux: &linux.Options{
			ProgramName: "novella",
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
