import { useState, useEffect } from "react";
import {
  GetConfig,
  UpdateConfig,
  SelectDirectory,
  ClearCache,
  GetCacheInfo,
} from "../../../wailsjs/go/main/App";
import type { CacheInfo } from "../../types";

export function SettingsTab() {
  const [workerCount, setWorkerCount] = useState(4);
  const [chunkSize, setChunkSize] = useState(3500);
  const [outputDir, setOutputDir] = useState("");
  const [theme, setTheme] = useState("dark");
  const [outputFormat, setOutputFormat] = useState("txt");
  const [maxRetries, setMaxRetries] = useState(2);
  const [enableReview, setEnableReview] = useState(true);
  const [cacheInfo, setCacheInfo] = useState<CacheInfo | null>(null);

  useEffect(() => {
    loadSettings();
    loadCacheInfo();
  }, []);

  const loadCacheInfo = async () => {
    try {
      const info = await GetCacheInfo();
      setCacheInfo(info);
    } catch (e) {
      console.error(e);
    }
  };

  const handleClearCache = async () => {
    try {
      await ClearCache();
      loadCacheInfo();
    } catch (e) {
      console.error(e);
    }
  };

  const loadSettings = async () => {
    try {
      const config = await GetConfig();
      setWorkerCount(config.workerCount || 4);
      setChunkSize(config.chunkSize || 3500);
      setOutputDir(config.outputDir || "");
      setTheme(config.theme || "dark");
      setOutputFormat(config.outputFormat || "txt");
      setMaxRetries(config.maxRetries ?? 2);
      setEnableReview(config.enableReview ?? true);
    } catch (e) {
      console.error(e);
    }
  };

  const saveSetting = async (key: string, value: any) => {
    try {
      await UpdateConfig({ [key]: value });
    } catch (e) {
      console.error(e);
    }
  };

  const handleSelectOutputDir = async () => {
    try {
      const dir = await SelectDirectory();
      if (dir) {
        setOutputDir(dir);
        saveSetting("outputDir", dir);
      }
    } catch (e) {
      console.error(e);
    }
  };

  return (
    <div className="space-y-4">
      <div className="space-y-1 mb-6">
        <h2 className="text-lg font-light tracking-tight">Settings</h2>
        <p className="text-xs text-muted-foreground">Translation preferences</p>
      </div>

      <div className="space-y-5">
        <div className="space-y-2">
          <label className="text-xs text-muted-foreground uppercase tracking-wider">
            Workers
          </label>
          <div className="flex items-center gap-4">
            <input
              type="range"
              min="1"
              max="20"
              value={workerCount}
              onChange={(e) => {
                setWorkerCount(Number(e.target.value));
                saveSetting("workerCount", Number(e.target.value));
              }}
              className="flex-1 accent-color-foreground"
            />
            <span className="text-sm w-6 text-right">{workerCount}</span>
          </div>
        </div>

        <div className="space-y-2">
          <label className="text-xs text-muted-foreground uppercase tracking-wider">
            Max Retries
          </label>
          <div className="flex items-center gap-4">
            <input
              type="range"
              min="0"
              max="5"
              value={maxRetries}
              onChange={(e) => {
                setMaxRetries(Number(e.target.value));
                saveSetting("maxRetries", Number(e.target.value));
              }}
              className="flex-1 accent-color-foreground"
            />
            <span className="text-sm w-6 text-right">{maxRetries}</span>
          </div>
        </div>

        <div className="space-y-2">
          <label className="text-xs text-muted-foreground uppercase tracking-wider">
            Chunk Size
          </label>
          <input
            type="number"
            value={chunkSize}
            onChange={(e) => {
              setChunkSize(Number(e.target.value));
              saveSetting("chunkSize", Number(e.target.value));
            }}
            className="w-32 h-8 px-3 rounded-md border border-border bg-transparent text-sm focus:outline-none focus:ring-1 focus:ring-foreground/20"
          />
        </div>

        <div className="space-y-2">
          <label className="text-xs text-muted-foreground uppercase tracking-wider">
            Output Format
          </label>
          <div className="flex gap-2">
            {["txt", "md", "epub"].map((fmt) => (
              <button
                key={fmt}
                onClick={() => {
                  setOutputFormat(fmt);
                  saveSetting("outputFormat", fmt);
                }}
                className={`px-3 h-8 rounded-md text-xs uppercase transition-colors ${
                  outputFormat === fmt
                    ? "bg-foreground text-background"
                    : "border border-border text-muted-foreground hover:text-foreground"
                }`}
              >
                {fmt}
              </button>
            ))}
          </div>
        </div>

        <div className="space-y-2">
          <label className="text-xs text-muted-foreground uppercase tracking-wider">
            Review Step
          </label>
          <button
            onClick={() => {
              setEnableReview(!enableReview);
              saveSetting("enableReview", !enableReview);
            }}
            className={`w-8 h-4 rounded-full transition-colors ${
              enableReview ? "bg-foreground" : "bg-muted"
            }`}
          >
            <div
              className={`w-3 h-3 rounded-full bg-background transition-transform ${
                enableReview ? "translate-x-4" : "translate-x-0.5"
              }`}
            />
          </button>
        </div>

        <div className="space-y-2">
          <label className="text-xs text-muted-foreground uppercase tracking-wider">
            Output Directory
          </label>
          <div className="flex gap-2">
            <input
              type="text"
              value={outputDir}
              readOnly
              placeholder="Select output directory..."
              className="flex-1 h-8 px-3 rounded-md border border-border bg-transparent text-sm truncate"
            />
            <button
              onClick={handleSelectOutputDir}
              className="h-8 px-3 rounded-md text-xs border border-border hover:bg-foreground/5 transition-colors"
            >
              Browse
            </button>
          </div>
        </div>

        <div className="space-y-2">
          <label className="text-xs text-muted-foreground uppercase tracking-wider">
            Theme
          </label>
          <div className="flex gap-2">
            {["light", "dark", "system"].map((t) => (
              <button
                key={t}
                onClick={() => {
                  setTheme(t);
                  saveSetting("theme", t);
                  document.documentElement.classList.toggle(
                    "dark",
                    t === "dark",
                  );
                }}
                className={`px-3 h-8 rounded-md text-xs capitalize transition-colors ${
                  theme === t
                    ? "bg-foreground text-background"
                    : "border border-border text-muted-foreground hover:text-foreground"
                }`}
              >
                {t}
              </button>
            ))}
          </div>
        </div>

        <div className="space-y-2 pt-2 border-t border-border">
          <label className="text-xs text-muted-foreground uppercase tracking-wider">
            Cache
          </label>
          <div className="flex items-center justify-between">
            <span className="text-xs text-muted-foreground">
              {cacheInfo ? `${cacheInfo.sizeHuman} cached` : "Loading..."}
            </span>
            <button
              onClick={handleClearCache}
              className="h-8 px-3 rounded-md text-xs border border-red-500/30 text-red-400 hover:bg-red-500/10 transition-colors"
            >
              Clear Cache
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}
