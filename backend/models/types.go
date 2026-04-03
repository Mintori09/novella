package models

type ProviderConfig struct {
	APIKey  string `json:"apiKey"`
	Model   string `json:"model"`
	Enabled bool   `json:"enabled"`
}

type AppConfig struct {
	Providers              map[string]ProviderConfig `json:"providers"`
	ActiveProvider         string                    `json:"activeProvider"`
	WorkerCount            int                       `json:"workerCount"`
	ChunkSize              int                       `json:"chunkSize"`
	EnableReview           bool                      `json:"enableReview"`
	OutputDir              string                    `json:"outputDir"`
	OutputFormat           string                    `json:"outputFormat"`
	FallbackOnFailure      bool                      `json:"fallbackOnFailure"`
	Theme                  string                    `json:"theme"`
	CustomPromptDir        string                    `json:"customPromptDir"`
	CustomPromptDirEnabled bool                      `json:"customPromptDirEnabled"`
	MaxRetries             int                       `json:"maxRetries"`
}

type TranslationTask struct {
	ID           string  `json:"id"`
	Name         string  `json:"name"`
	Genre        string  `json:"genre"`
	InputPath    string  `json:"inputPath"`
	OutputPath   string  `json:"outputPath"`
	Status       string  `json:"status"`
	TotalChunks  int     `json:"totalChunks"`
	DoneChunks   int     `json:"doneChunks"`
	CurrentChunk int     `json:"currentChunk"`
	Progress     float64 `json:"progress"`
	Error        string  `json:"error"`
	StartedAt    string  `json:"startedAt"`
	FinishedAt   string  `json:"finishedAt"`
}

type ChunkResult struct {
	Index    int    `json:"index"`
	Content  string `json:"content"`
	Error    string `json:"error"`
	IsReview bool   `json:"isReview"`
}

type StreamChunk struct {
	Index    int    `json:"index"`
	FileName string `json:"fileName"`
	Content  string `json:"content"`
	Done     bool   `json:"done"`
}

type TokenUsage struct {
	InputTokens  int `json:"inputTokens"`
	OutputTokens int `json:"outputTokens"`
	TotalTokens  int `json:"totalTokens"`
}

type TaskProgress struct {
	TaskID         string     `json:"taskId"`
	FileName       string     `json:"fileName"`
	Status         string     `json:"status"`
	CurrentChunk   int        `json:"currentChunk"`
	TotalChunks    int        `json:"totalChunks"`
	ProcessedChars int        `json:"processedChars"`
	TotalChars     int        `json:"totalChars"`
	IsReviewing    bool       `json:"isReviewing"`
	Progress       float64    `json:"progress"`
	Error          string     `json:"error"`
	TokenUsage     TokenUsage `json:"tokenUsage"`
}

type GlossaryEntry struct {
	Original   string `json:"original"`
	Translated string `json:"translated"`
	Category   string `json:"category"`
	Note       string `json:"note"`
}

type Glossary struct {
	NovelName string                   `json:"novelName"`
	Author    string                   `json:"author"`
	Genre     string                   `json:"genre"`
	Entries   map[string]GlossaryEntry `json:"entries"`
}

type APIProvider struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	BaseURL      string `json:"baseUrl"`
	DefaultModel string `json:"defaultModel"`
	APIKeyEnv    string `json:"apiKeyEnv"`
}

type TestConnectionResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type FileInfo struct {
	Name  string `json:"name"`
	Path  string `json:"path"`
	Size  int64  `json:"size"`
	IsDir bool   `json:"isDir"`
}

type CacheEntry struct {
	ContentHash       string `json:"contentHash"`
	Genre             string `json:"genre"`
	OutputPath        string `json:"outputPath"`
	TranslatedContent string `json:"translatedContent"`
	Timestamp         string `json:"timestamp"`
}

type GenrePrompt struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Content string `json:"content"`
	Builtin bool   `json:"builtin"`
}

type EPUBMetadata struct {
	Title       string
	Author      string
	Description string
	CoverPath   string
	Language    string
}
