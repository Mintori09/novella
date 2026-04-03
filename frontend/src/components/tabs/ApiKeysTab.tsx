import { useState, useEffect } from "react";
import {
  GetConfig,
  UpdateConfig,
  GetProviders,
  TestConnection,
} from "../../../wailsjs/go/main/App";
import type { ProviderConfig, APIProvider } from "../../types";

export function ApiKeysTab() {
  const [providers, setProviders] = useState<Record<string, ProviderConfig>>(
    {},
  );
  const [apiProviders, setApiProviders] = useState<Record<string, APIProvider>>(
    {},
  );
  const [activeProvider, setActiveProvider] = useState("kilo");
  const [testing, setTesting] = useState<string | null>(null);
  const [testResults, setTestResults] = useState<
    Record<string, { success: boolean; message: string }>
  >({});

  useEffect(() => {
    loadConfig();
  }, []);

  const loadConfig = async () => {
    try {
      const config = await GetConfig();
      setProviders(config.providers || {});
      setActiveProvider(config.activeProvider || "kilo");
    } catch (e) {
      console.error(e);
    }
    try {
      const ap = await GetProviders();
      setApiProviders(ap);
    } catch (e) {
      console.error(e);
    }
  };

  const handleApiKeyChange = async (id: string, apiKey: string) => {
    const updated = { ...providers, [id]: { ...providers[id], apiKey } };
    setProviders(updated);
    await UpdateConfig({ providers: updated });
  };

  const handleModelChange = async (id: string, model: string) => {
    const updated = { ...providers, [id]: { ...providers[id], model } };
    setProviders(updated);
    await UpdateConfig({ providers: updated });
  };

  const handleToggleEnabled = async (id: string) => {
    const updated = {
      ...providers,
      [id]: { ...providers[id], enabled: !providers[id].enabled },
    };
    setProviders(updated);
    await UpdateConfig({ providers: updated });
  };

  const handleSetActive = async (id: string) => {
    setActiveProvider(id);
    await UpdateConfig({ activeProvider: id });
  };

  const handleTest = async (id: string) => {
    const config = providers[id];
    if (!config?.apiKey?.trim()) {
      setTestResults((prev) => ({
        ...prev,
        [id]: { success: false, message: "API key is required" },
      }));
      return;
    }
    if (!config?.model?.trim()) {
      setTestResults((prev) => ({
        ...prev,
        [id]: { success: false, message: "Model is required" },
      }));
      return;
    }
    setTesting(id);
    try {
      const result = await TestConnection(id, config.apiKey);
      setTestResults((prev) => ({ ...prev, [id]: result }));
    } catch (e: any) {
      setTestResults((prev) => ({
        ...prev,
        [id]: { success: false, message: e.message || String(e) },
      }));
    } finally {
      setTesting(null);
    }
  };

  return (
    <div className="space-y-4">
      <div className="space-y-1 mb-6">
        <h2 className="text-lg font-light tracking-tight">API Keys</h2>
        <p className="text-xs text-muted-foreground">
          Configure providers and models
        </p>
      </div>

      <div className="space-y-3">
        {Object.entries(apiProviders).map(([id, provider]) => {
          const config = providers[id] || {
            apiKey: "",
            model: provider.defaultModel,
            enabled: false,
          };
          const isTested = testResults[id];

          return (
            <div
              key={id}
              className={`p-4 rounded-lg border transition-colors ${
                activeProvider === id ? "border-foreground/30" : "border-border"
              }`}
            >
              <div className="flex items-center justify-between mb-3">
                <div className="flex items-center gap-3">
                  <button
                    onClick={() => handleSetActive(id)}
                    className={`w-3.5 h-3.5 rounded-full border-2 transition-colors ${
                      activeProvider === id
                        ? "border-foreground bg-foreground"
                        : "border-muted-foreground"
                    }`}
                  />
                  <span className="text-sm font-medium">{provider.name}</span>
                </div>
                <button
                  onClick={() => handleToggleEnabled(id)}
                  className={`w-8 h-4 rounded-full transition-colors ${
                    config.enabled ? "bg-foreground" : "bg-muted"
                  }`}
                >
                  <div
                    className={`w-3 h-3 rounded-full bg-background transition-transform ${
                      config.enabled ? "translate-x-4" : "translate-x-0.5"
                    }`}
                  />
                </button>
              </div>

              <div className="space-y-2">
                <input
                  type="password"
                  value={config.apiKey}
                  onChange={(e) => handleApiKeyChange(id, e.target.value)}
                  placeholder={`API Key`}
                  className="w-full h-8 px-3 rounded-md border border-border bg-transparent text-sm placeholder:text-muted-foreground/50 focus:outline-none focus:ring-1 focus:ring-foreground/20"
                />
                <div className="flex gap-2">
                  <input
                    type="text"
                    value={config.model}
                    onChange={(e) => handleModelChange(id, e.target.value)}
                    placeholder="Model ID"
                    className="flex-1 h-8 px-3 rounded-md border border-border bg-transparent text-sm placeholder:text-muted-foreground/50 focus:outline-none focus:ring-1 focus:ring-foreground/20 font-mono"
                  />
                  <button
                    onClick={() => handleTest(id)}
                    disabled={testing === id}
                    className="h-8 px-3 rounded-md text-xs border border-border hover:bg-foreground/5 transition-colors disabled:opacity-50"
                  >
                    {testing === id ? "..." : "Test"}
                  </button>
                </div>
              </div>

              {isTested && (
                <p
                  className={`text-xs mt-2 ${isTested.success ? "text-green-500" : "text-red-500"}`}
                >
                  {isTested.message}
                </p>
              )}
            </div>
          );
        })}
      </div>
    </div>
  );
}
