package engine

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"novella/backend/cache"
	"novella/backend/config"
	"novella/backend/exporter"
	"novella/backend/glossary"
	"novella/backend/models"
	promptsmgr "novella/backend/prompts"
	"novella/backend/provider"
	"novella/prompts"

	"github.com/google/uuid"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type Engine struct {
	cfgManager    *config.Manager
	chunker       *Chunker
	glossaryMgr   *glossary.Manager
	cacheMgr      *cache.Manager
	promptMgr     *promptsmgr.Manager
	provider      provider.Provider
	mu            sync.Mutex
	tasks         map[string]*models.TranslationTask
	cancelFuncs   map[string]context.CancelFunc
	activeWorkers atomic.Int64
	maxWorkers    atomic.Int64
}

func NewEngine(cfgManager *config.Manager, glossaryMgr *glossary.Manager, cacheMgr *cache.Manager, promptMgr *promptsmgr.Manager) *Engine {
	maxW := cfgManager.Get().WorkerCount
	if maxW < 1 {
		maxW = 1
	}
	if maxW > 10 {
		maxW = 10
	}
	e := &Engine{
		cfgManager:  cfgManager,
		chunker:     NewChunker(cfgManager.Get().ChunkSize),
		glossaryMgr: glossaryMgr,
		cacheMgr:    cacheMgr,
		promptMgr:   promptMgr,
		tasks:       make(map[string]*models.TranslationTask),
		cancelFuncs: make(map[string]context.CancelFunc),
	}
	e.maxWorkers.Store(int64(maxW))
	return e
}

func (e *Engine) UpdateMaxWorkers(count int) {
	if count < 1 {
		count = 1
	}
	if count > 10 {
		count = 10
	}
	e.maxWorkers.Store(int64(count))
}

func calcProgress(processedChars, totalChars int, isReviewing bool, enableReview bool) float64 {
	if totalChars == 0 {
		return 0
	}
	translatePct := float64(processedChars) / float64(totalChars)
	if !enableReview {
		return translatePct * 100
	}
	if !isReviewing {
		return translatePct * 70
	}
	return 70 + (translatePct * 30)
}

var rateLimitRegex = regexp.MustCompile(`(?i)(?:try again in|retry after)[\s:]*([\d.]+)s?`)

func isRateLimit(err error) bool {
	msg := err.Error()
	return strings.Contains(msg, "429") || strings.Contains(msg, "rate limit") || strings.Contains(msg, "rate_limit")
}

func getRateLimitDelay(err error) time.Duration {
	msg := err.Error()
	if matches := rateLimitRegex.FindStringSubmatch(msg); len(matches) > 1 {
		if seconds, parseErr := strconv.ParseFloat(matches[1], 64); parseErr == nil {
			return time.Duration(seconds*1000+500) * time.Millisecond
		}
	}
	return 10 * time.Second
}

func (e *Engine) loadPrompt(genre string) (systemPrompt string, genrePrompt string, err error) {
	cfg := e.cfgManager.Get()

	if cfg.CustomPromptDirEnabled && cfg.CustomPromptDir != "" {
		systemPromptBytes, err := os.ReadFile(filepath.Join(cfg.CustomPromptDir, "system_prompt.md"))
		if err != nil {
			return "", "", err
		}
		systemPrompt = string(systemPromptBytes)
		genrePromptBytes, err := os.ReadFile(filepath.Join(cfg.CustomPromptDir, genre+".md"))
		if err != nil {
			return "", "", err
		}
		genrePrompt = string(genrePromptBytes)
	} else {
		systemPrompt = e.promptMgr.GetSystemPrompt()
		if systemPrompt == "" {
			data, err := prompts.FS.ReadFile("system_prompt.md")
			if err != nil {
				return "", "", err
			}
			systemPrompt = string(data)
		}
		genrePrompt, err = e.promptMgr.GetGenrePrompt(genre)
		if err != nil {
			return "", "", err
		}
	}

	return systemPrompt, genrePrompt, nil
}

func (e *Engine) getProvider() (provider.Provider, models.ProviderConfig, error) {
	cfg := e.cfgManager.Get()
	activeID := cfg.ActiveProvider

	provCfg, ok := cfg.Providers[activeID]
	if !ok || !provCfg.Enabled {
		return nil, provCfg, fmt.Errorf("active provider '%s' not configured or disabled", activeID)
	}

	switch activeID {
	case "kilo":
		p := provider.NewKiloProvider()
		e.provider = p
		return p, provCfg, nil
	case "openrouter":
		p := provider.NewOpenRouterProvider()
		e.provider = p
		return p, provCfg, nil
	case "openai":
		p := provider.NewOpenAIProvider()
		e.provider = p
		return p, provCfg, nil
	case "groq":
		p := provider.NewGroqProvider()
		e.provider = p
		return p, provCfg, nil
	case "gemini":
		p := provider.NewGeminiProvider()
		e.provider = p
		return p, provCfg, nil
	case "nanogpt":
		p := provider.NewNanoGPTProvider()
		e.provider = p
		return p, provCfg, nil
	case "custom":
		if provCfg.CustomURL == "" {
			return nil, provCfg, fmt.Errorf("custom provider requires a URL")
		}
		method := provCfg.Method
		if method == "" {
			method = "POST"
		}
		p := provider.NewCustomProvider(provCfg.CustomURL, method)
		e.provider = p
		return p, provCfg, nil
	default:
		return nil, provCfg, fmt.Errorf("unknown provider: %s", activeID)
	}
}

func (e *Engine) emitProgress(ctx context.Context, taskID, fileName, status string, currentChunk, totalChunks, processedChars, totalChars int, isReviewing bool, progress float64, errMsg string) {
	e.emitProgressWithTokens(ctx, taskID, fileName, status, currentChunk, totalChunks, processedChars, totalChars, isReviewing, progress, errMsg, models.TokenUsage{})
}

func (e *Engine) emitProgressWithTokens(ctx context.Context, taskID, fileName, status string, currentChunk, totalChunks, processedChars, totalChars int, isReviewing bool, progress float64, errMsg string, tokenUsage models.TokenUsage) {
	runtime.EventsEmit(ctx, "task:progress", models.TaskProgress{
		TaskID:         taskID,
		FileName:       fileName,
		Status:         status,
		CurrentChunk:   currentChunk,
		TotalChunks:    totalChunks,
		ProcessedChars: processedChars,
		TotalChars:     totalChars,
		IsReviewing:    isReviewing,
		Progress:       progress,
		Error:          errMsg,
		TokenUsage:     tokenUsage,
	})
}

func (e *Engine) translateSingleFile(ctx context.Context, inputPath string, genre string) {
	taskID := uuid.New().String()
	fileName := filepath.Base(inputPath)
	cfg := e.cfgManager.Get()
	enableReview := cfg.EnableReview
	maxRetries := cfg.MaxRetries
	if maxRetries < 0 {
		maxRetries = 0
	}
	if maxRetries > 5 {
		maxRetries = 5
	}

	content, err := os.ReadFile(inputPath)
	if err != nil {
		e.emitProgress(ctx, taskID, fileName, "error", 0, 0, 0, 0, false, 0, err.Error())
		return
	}

	text := string(content)

	task := &models.TranslationTask{
		ID:          taskID,
		Name:        fileName,
		Genre:       genre,
		InputPath:   inputPath,
		Status:      "running",
		TotalChunks: 0,
		StartedAt:   time.Now().Format(time.RFC3339),
	}

	if cached, ok := e.cacheMgr.Check(text, genre); ok {
		outputDir := cfg.OutputDir
		if outputDir == "" {
			outputDir = filepath.Dir(inputPath)
		}
		os.MkdirAll(outputDir, 0o755)

		baseName := strings.TrimSuffix(filepath.Base(inputPath), filepath.Ext(inputPath))
		outputFormat := cfg.OutputFormat
		if outputFormat == "" {
			outputFormat = "txt"
		}

		meta := extractMetadata(inputPath, text, genre)
		exp := exporter.NewExporter(outputFormat)
		outputPath := filepath.Join(outputDir, baseName)
		finalPath, err := exp.Export(cached.TranslatedContent, baseName, outputPath, meta)
		if err != nil {
			task.Status = "error"
			task.Error = err.Error()
			e.emitProgress(ctx, taskID, fileName, "error", 0, 0, 0, 0, false, 0, err.Error())
			return
		}

		task.Status = "completed"
		task.Progress = 100
		task.DoneChunks = 0
		task.OutputPath = finalPath
		task.FinishedAt = time.Now().Format(time.RFC3339)
		e.mu.Lock()
		e.tasks[taskID] = task
		e.mu.Unlock()
		e.emitProgress(ctx, taskID, fileName, "cached", 0, 0, 0, 0, false, 100, "")
		return
	}

	chunks := e.chunker.Split(text)

	totalChars := 0
	for _, c := range chunks {
		totalChars += len(c)
	}

	task.TotalChunks = len(chunks)

	taskCtx, cancel := context.WithCancel(ctx)
	e.mu.Lock()
	e.tasks[taskID] = task
	e.cancelFuncs[taskID] = cancel
	e.mu.Unlock()

	runtime.EventsEmit(ctx, "task:started", map[string]any{
		"taskId":   taskID,
		"fileName": fileName,
		"genre":    genre,
	})

	prov, provCfg, err := e.getProvider()
	if err != nil {
		task.Status = "error"
		task.Error = err.Error()
		e.emitProgress(ctx, taskID, fileName, "error", 0, len(chunks), 0, totalChars, false, 0, err.Error())
		return
	}

	systemPrompt, genrePrompt, err := e.loadPrompt(genre)
	if err != nil {
		task.Status = "error"
		task.Error = err.Error()
		e.emitProgress(ctx, taskID, fileName, "error", 0, len(chunks), 0, totalChars, false, 0, err.Error())
		return
	}

	glossaryPrompt := e.glossaryMgr.ToPromptString(taskID)

	if len(e.glossaryMgr.Get(taskID).Entries) == 0 && len(text) > 500 {
		e.emitProgress(ctx, taskID, fileName, "extracting_glossary", 0, 0, 0, totalChars, false, 0, "")

		extractPrompt := buildGlossaryExtractionPrompt(text[:min(len(text), 8000)], genre)
		genCfg := provider.GenerateConfig{
			SystemPrompt: "You are a terminology extraction assistant. Return only the glossary list, no explanations.",
			Temperature:  0.1,
			MaxTokens:    4096,
			Model:        provCfg.Model,
			APIKey:       provCfg.APIKey,
		}

		result, err := prov.GenerateContent(taskCtx, extractPrompt, genCfg)
		if err == nil {
			terms := glossary.ParseExtractedTerms(result.Content)
			if len(terms) > 0 {
				e.glossaryMgr.MergeEntries(taskID, terms)
				glossaryPrompt = e.glossaryMgr.ToPromptString(taskID)
			}
		}
	}

	var translatedChunks []string
	var processedChars int
	var streamContent string
	var totalTokenUsage models.TokenUsage

	for i, chunk := range chunks {
		select {
		case <-taskCtx.Done():
			task.Status = "cancelled"
			e.emitProgress(ctx, taskID, fileName, "cancelled", i, len(chunks), processedChars, totalChars, false, calcProgress(processedChars, totalChars, false, enableReview), "")
			return
		default:
		}

		processedChars += len(chunk)
		progress := calcProgress(processedChars, totalChars, false, enableReview)

		e.emitProgress(ctx, taskID, fileName, "translating", i+1, len(chunks), processedChars, totalChars, false, progress, "")

		prompt := buildTranslationPrompt(systemPrompt, genrePrompt, glossaryPrompt, chunk, streamContent)

		genCfg := provider.GenerateConfig{
			SystemPrompt: "You are a professional literary translator. Return only the translated text, no explanations.",
			Temperature:  0.3,
			MaxTokens:    8192,
			Model:        provCfg.Model,
			APIKey:       provCfg.APIKey,
		}

		var result provider.GenerateResult
		var lastErr error
		for attempt := 0; attempt <= maxRetries; attempt++ {
			select {
			case <-taskCtx.Done():
				task.Status = "cancelled"
				e.emitProgress(ctx, taskID, fileName, "cancelled", i, len(chunks), processedChars, totalChars, false, calcProgress(processedChars, totalChars, false, enableReview), "")
				return
			default:
			}

			result, err = prov.GenerateContent(taskCtx, prompt, genCfg)
			if err == nil {
				break
			}

			lastErr = err
			if attempt < maxRetries {
				if isRateLimit(err) {
					delay := getRateLimitDelay(err)
					e.emitProgress(ctx, taskID, fileName, "retrying", i+1, len(chunks), processedChars, totalChars, false, progress, fmt.Sprintf("rate limited, waiting %v", delay.Round(time.Second)))
					time.Sleep(delay)
				} else {
					backoff := time.Duration(1<<uint(attempt)) * time.Second
					e.emitProgress(ctx, taskID, fileName, "retrying", i+1, len(chunks), processedChars, totalChars, false, progress, fmt.Sprintf("retry %d/%d", attempt+1, maxRetries))
					time.Sleep(backoff)
				}
			}
		}

		if lastErr != nil {
			task.Status = "error"
			task.Error = lastErr.Error()
			e.emitProgress(ctx, taskID, fileName, "error", i+1, len(chunks), processedChars, totalChars, false, progress, lastErr.Error())
			return
		}

		totalTokenUsage.InputTokens += result.TokenUsage.InputTokens
		totalTokenUsage.OutputTokens += result.TokenUsage.OutputTokens
		totalTokenUsage.TotalTokens += result.TokenUsage.TotalTokens

		translatedChunks = append(translatedChunks, result.Content)
		streamContent = result.Content
		if len(streamContent) > 500 {
			streamContent = streamContent[len(streamContent)-500:]
		}

		runtime.EventsEmit(ctx, "chunk:complete", models.ChunkResult{
			Index:   i,
			Content: result.Content,
		})

		time.Sleep(300 * time.Millisecond)
	}

	// Review step
	if enableReview && len(translatedChunks) > 0 {
		e.emitProgress(ctx, taskID, fileName, "reviewing", len(chunks), len(chunks), totalChars, totalChars, true, 70, "")

		fullTranslation := strings.Join(translatedChunks, "\n\n")
		reviewPrompt := buildReviewPrompt(genre, fullTranslation)

		genCfg := provider.GenerateConfig{
			SystemPrompt: "You are a professional literary translator. Return only the polished translation, no explanations.",
			Temperature:  0.3,
			MaxTokens:    16384,
			Model:        provCfg.Model,
			APIKey:       provCfg.APIKey,
		}

		var lastErr error
		for attempt := 0; attempt <= maxRetries; attempt++ {
			select {
			case <-taskCtx.Done():
				task.Status = "cancelled"
				e.emitProgress(ctx, taskID, fileName, "cancelled", len(chunks), len(chunks), totalChars, totalChars, true, 70, "")
				return
			default:
			}

			reviewed, err := prov.GenerateContent(taskCtx, reviewPrompt, genCfg)
			if err == nil {
				translatedChunks = []string{reviewed.Content}
				totalTokenUsage.InputTokens += reviewed.TokenUsage.InputTokens
				totalTokenUsage.OutputTokens += reviewed.TokenUsage.OutputTokens
				totalTokenUsage.TotalTokens += reviewed.TokenUsage.TotalTokens
				break
			}
			lastErr = err
			if attempt < maxRetries {
				if isRateLimit(err) {
					delay := getRateLimitDelay(err)
					time.Sleep(delay)
				} else {
					backoff := time.Duration(1<<uint(attempt)) * time.Second
					time.Sleep(backoff)
				}
			}
		}

		if lastErr != nil {
			// Review failed, keep original translation
			fmt.Fprintf(os.Stderr, "review failed: %v\n", lastErr)
		}
	}

	if task.Status == "running" {
		task.Status = "completed"
		task.DoneChunks = len(chunks)
		task.Progress = 100
		task.FinishedAt = time.Now().Format(time.RFC3339)

		output := strings.Join(translatedChunks, "\n\n")

		outputDir := cfg.OutputDir
		if outputDir == "" {
			outputDir = filepath.Dir(inputPath)
		}
		os.MkdirAll(outputDir, 0o755)

		baseName := strings.TrimSuffix(filepath.Base(inputPath), filepath.Ext(inputPath))
		outputFormat := cfg.OutputFormat
		if outputFormat == "" {
			outputFormat = "txt"
		}

		meta := extractMetadata(inputPath, text, genre)
		exp := exporter.NewExporter(outputFormat)
		outputPath := filepath.Join(outputDir, baseName)
		finalPath, err := exp.Export(output, baseName, outputPath, meta)
		if err != nil {
			task.Status = "error"
			task.Error = err.Error()
			e.emitProgress(ctx, taskID, fileName, "error", len(chunks), len(chunks), totalChars, totalChars, false, 100, err.Error())
			return
		}

		task.OutputPath = finalPath

		contentHash := e.cacheMgr.HashContent(text)
		e.cacheMgr.Save(contentHash, genre, output, finalPath, time.Now().Format(time.RFC3339))

		e.emitProgressWithTokens(ctx, taskID, fileName, "completed", len(chunks), len(chunks), totalChars, totalChars, false, 100, "", totalTokenUsage)
	}
}

func (e *Engine) TranslateFiles(ctx context.Context, paths []string, genre string) {
	var wg sync.WaitGroup

	for _, path := range paths {
		for e.activeWorkers.Load() >= e.maxWorkers.Load() {
			time.Sleep(10 * time.Millisecond)
		}
		e.activeWorkers.Add(1)
		wg.Add(1)
		go func(p string) {
			defer wg.Done()
			defer e.activeWorkers.Add(-1)
			e.translateSingleFile(ctx, p, genre)
		}(path)
	}
	wg.Wait()
}

func (e *Engine) TranslateFile(ctx context.Context, inputPath string, genre string) {
	e.translateSingleFile(ctx, inputPath, genre)
}

func (e *Engine) CancelTask(taskID string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if cancel, ok := e.cancelFuncs[taskID]; ok {
		cancel()
		delete(e.cancelFuncs, taskID)
	}
}

func (e *Engine) StopAll() {
	e.mu.Lock()
	defer e.mu.Unlock()
	for taskID, cancel := range e.cancelFuncs {
		cancel()
		delete(e.cancelFuncs, taskID)
	}
}

func (e *Engine) GetTasks() []*models.TranslationTask {
	e.mu.Lock()
	defer e.mu.Unlock()
	var tasks []*models.TranslationTask
	for _, t := range e.tasks {
		tasks = append(tasks, t)
	}
	return tasks
}

func (e *Engine) TestProviderConnection(ctx context.Context, providerID string, apiKey string, model string) (*models.TestConnectionResult, error) {
	var p provider.Provider
	switch providerID {
	case "kilo":
		p = provider.NewKiloProvider()
	case "openrouter":
		p = provider.NewOpenRouterProvider()
	case "openai":
		p = provider.NewOpenAIProvider()
	case "groq":
		p = provider.NewGroqProvider()
	case "gemini":
		p = provider.NewGeminiProvider()
	case "nanogpt":
		p = provider.NewNanoGPTProvider()
	case "custom":
		if provCfg, ok := e.cfgManager.Get().Providers["custom"]; ok && provCfg.CustomURL != "" {
			method := provCfg.Method
			if method == "" {
				method = "POST"
			}
			p = provider.NewCustomProvider(provCfg.CustomURL, method)
		} else {
			return &models.TestConnectionResult{Success: false, Message: "Custom provider URL not configured"}, nil
		}
	default:
		return &models.TestConnectionResult{Success: false, Message: "Unknown provider"}, nil
	}

	err := p.TestConnection(ctx, apiKey, model)
	if err != nil {
		return &models.TestConnectionResult{Success: false, Message: err.Error()}, nil
	}

	return &models.TestConnectionResult{Success: true, Message: "Connection successful"}, nil
}

func buildTranslationPrompt(systemPrompt, genrePrompt, glossaryPrompt, chunk, context string) string {
	var parts []string
	parts = append(parts, systemPrompt)
	parts = append(parts, "")
	parts = append(parts, genrePrompt)
	parts = append(parts, "")
	parts = append(parts, "## Glossary:")
	if glossaryPrompt != "" {
		parts = append(parts, glossaryPrompt)
	} else {
		parts = append(parts, "(No glossary yet)")
	}
	if context != "" {
		parts = append(parts, "")
		parts = append(parts, "## Previous context:")
		parts = append(parts, context)
	}
	parts = append(parts, "")
	parts = append(parts, "## Translate this:")
	parts = append(parts, chunk)
	parts = append(parts, "")
	parts = append(parts, "Return only the Vietnamese translation, no explanations.")

	return strings.Join(parts, "\n")
}

func buildReviewPrompt(genre, translation string) string {
	return strings.Join([]string{
		"Review and polish this Vietnamese translation.",
		"Fix awkward phrasing, ensure natural flow, check consistency.",
		"Return only the polished translation.",
		"",
		"Translation to review:",
		translation,
	}, "\n")
}

func buildGlossaryExtractionPrompt(sampleText string, genre string) string {
	return strings.Join([]string{
		"Extract important terminology from this text for a translation glossary.",
		"Focus on: character names, place names, special terms, titles, organizations.",
		"Format each entry as: - Original Term → Vietnamese Translation (optional note)",
		"Group by category: ### Names, ### Places, ### Terms, ### Titles",
		"Genre: " + genre,
		"",
		"Text sample:",
		sampleText,
		"",
		"Return only the formatted glossary list, no explanations.",
	}, "\n")
}

func extractMetadata(inputPath string, content string, genre string) *models.EPUBMetadata {
	meta := &models.EPUBMetadata{
		Title:    strings.TrimSuffix(filepath.Base(inputPath), filepath.Ext(inputPath)),
		Author:   "Unknown",
		Language: "vi",
	}

	lines := strings.SplitN(content, "\n", 20)
	for _, line := range lines {
		line = strings.TrimSpace(line)
		lower := strings.ToLower(line)

		if strings.HasPrefix(lower, "title:") || strings.HasPrefix(lower, "tiêu đề:") {
			meta.Title = strings.TrimSpace(line[strings.Index(line, ":")+1:])
		} else if strings.HasPrefix(lower, "author:") || strings.HasPrefix(lower, "tác giả:") {
			meta.Author = strings.TrimSpace(line[strings.Index(line, ":")+1:])
		} else if strings.HasPrefix(lower, "description:") || strings.HasPrefix(lower, "mô tả:") {
			meta.Description = strings.TrimSpace(line[strings.Index(line, ":")+1:])
		}
	}

	coverPath := filepath.Join(filepath.Dir(inputPath), "cover.jpg")
	if _, err := os.Stat(coverPath); err == nil {
		meta.CoverPath = coverPath
	} else {
		for _, ext := range []string{".png", ".jpeg", ".gif", ".webp"} {
			coverPath = filepath.Join(filepath.Dir(inputPath), "cover"+ext)
			if _, err := os.Stat(coverPath); err == nil {
				meta.CoverPath = coverPath
				break
			}
		}
	}

	return meta
}
