export interface ProviderConfig {
  apiKey: string;
  model: string;
  models: string[];
  defaultModel: string;
  enabled: boolean;
  customUrl: string;
  method: string;
}

export interface AppConfig {
  providers: Record<string, ProviderConfig>;
  activeProvider: string;
  workerCount: number;
  chunkSize: number;
  enableReview: boolean;
  outputDir: string;
  outputFormat: string;
  fallbackOnFailure: boolean;
  theme: string;
  customPromptDir: string;
  customPromptDirEnabled: boolean;
  maxRetries: number;
}

export interface TranslationTask {
  id: string;
  name: string;
  genre: string;
  inputPath: string;
  outputPath: string;
  status: string;
  totalChunks: number;
  doneChunks: number;
  currentChunk: number;
  progress: number;
  error: string;
  startedAt: string;
  finishedAt: string;
}

export interface ChunkResult {
  index: number;
  content: string;
  error: string;
  isReview: boolean;
}

export interface TaskProgress {
  taskId: string;
  fileName: string;
  status: string;
  currentChunk: number;
  totalChunks: number;
  processedChars: number;
  totalChars: number;
  isReviewing: boolean;
  progress: number;
  error: string;
  tokenUsage: {
    inputTokens: number;
    outputTokens: number;
    totalTokens: number;
  };
}

export interface StreamChunk {
  index: number;
  fileName: string;
  content: string;
  done: boolean;
}

export interface GlossaryEntry {
  original: string;
  translated: string;
  category: string;
  note: string;
}

export interface Glossary {
  novelName: string;
  author: string;
  genre: string;
  entries: Record<string, GlossaryEntry>;
}

export interface APIProvider {
  id: string;
  name: string;
  baseUrl: string;
  defaultModel: string;
  apiKeyEnv: string;
}

export interface TestConnectionResult {
  success: boolean;
  message: string;
}

export interface FileInfo {
  name: string;
  path: string;
  size: number;
  isDir: boolean;
}

export interface CacheInfo {
  size: number;
  sizeHuman: string;
}

export interface GenrePrompt {
  id: string;
  name: string;
  content: string;
  builtin: boolean;
}
