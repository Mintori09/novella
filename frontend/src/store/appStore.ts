import { create } from "zustand";
import type { AppConfig, FileInfo, Glossary } from "../types";

interface FileTask {
  taskId: string;
  fileName: string;
  status:
    | "waiting"
    | "translating"
    | "reviewing"
    | "done"
    | "error"
    | "cancelled"
    | "retrying"
    | "cached";
  progress: number;
  processedChars: number;
  totalChars: number;
  isReviewing: boolean;
  currentChunk: number;
  totalChunks: number;
  error?: string;
  tokenUsage?: {
    inputTokens: number;
    outputTokens: number;
    totalTokens: number;
  };
}

interface AppState {
  config: AppConfig | null;
  tasks: Record<string, FileTask>;
  streamContent: Record<string, string>;
  selectedGenre: string;
  glossaries: string[];
  currentGlossary: Glossary | null;
  activePage: string;
  theme: string;

  inputDir: string;
  inputDirValid: boolean;
  outputDir: string;
  files: FileInfo[];
  selectedFiles: string[];

  setConfig: (config: AppConfig) => void;
  patchConfig: (updates: Partial<AppConfig>) => void;
  setTask: (taskId: string, task: FileTask) => void;
  removeTask: (taskId: string) => void;
  clearTasks: () => void;
  setStreamContent: (taskId: string, content: string) => void;
  appendStreamContent: (taskId: string, chunk: string) => void;
  setSelectedGenre: (genre: string) => void;
  setGlossaries: (glossaries: string[]) => void;
  setCurrentGlossary: (glossary: Glossary | null) => void;
  setActivePage: (page: string) => void;
  setTheme: (theme: string) => void;

  setInputDir: (dir: string, valid: boolean) => void;
  setOutputDir: (dir: string) => void;
  setFiles: (files: FileInfo[]) => void;
  setSelectedFiles: (files: string[]) => void;
  toggleFile: (path: string) => void;
  selectAllFiles: () => void;
  deselectAllFiles: () => void;
}

export const useAppStore = create<AppState>((set, get) => ({
  config: null,
  tasks: {},
  streamContent: {},
  selectedGenre: "tien_hiep",
  glossaries: [],
  currentGlossary: null,
  activePage: "translate",
  theme: "dark",

  inputDir: "",
  inputDirValid: false,
  outputDir: "",
  files: [],
  selectedFiles: [],

  setConfig: (config) => set({ config }),

  patchConfig: (updates) =>
    set((state) => ({
      config: state.config ? { ...state.config, ...updates } : state.config,
    })),

  setTask: (taskId, task) =>
    set((state) => {
      const existing = state.tasks[taskId];
      if (existing && existing.status === "done") return state;
      if (existing && existing.status === "error" && task.status !== "error")
        return state;
      return { tasks: { ...state.tasks, [taskId]: task } };
    }),

  removeTask: (taskId) =>
    set((state) => {
      const next = { ...state.tasks };
      delete next[taskId];
      return { tasks: next };
    }),

  clearTasks: () => set({ tasks: {} }),

  setStreamContent: (taskId, content) =>
    set((state) => ({
      streamContent: { ...state.streamContent, [taskId]: content },
    })),

  appendStreamContent: (taskId, chunk) =>
    set((state) => ({
      streamContent: {
        ...state.streamContent,
        [taskId]: (state.streamContent[taskId] || "") + chunk,
      },
    })),

  setSelectedGenre: (genre) => set({ selectedGenre: genre }),
  setGlossaries: (glossaries) => set({ glossaries }),
  setCurrentGlossary: (glossary) => set({ currentGlossary: glossary }),
  setActivePage: (page) => set({ activePage: page }),
  setTheme: (theme) => set({ theme }),

  setInputDir: (dir, valid) => set({ inputDir: dir, inputDirValid: valid }),

  setOutputDir: (dir) => set({ outputDir: dir }),

  setFiles: (files) => set({ files }),

  setSelectedFiles: (files) => set({ selectedFiles: files }),

  toggleFile: (path) =>
    set((state) => {
      const next = state.selectedFiles.includes(path)
        ? state.selectedFiles.filter((p) => p !== path)
        : [...state.selectedFiles, path];
      return { selectedFiles: next };
    }),

  selectAllFiles: () =>
    set((state) => ({ selectedFiles: state.files.map((f) => f.path) })),

  deselectAllFiles: () => set({ selectedFiles: [] }),
}));
