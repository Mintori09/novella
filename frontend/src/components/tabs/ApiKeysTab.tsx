import { useState, useEffect, useRef } from "react";
import {
  GetConfig,
  UpdateConfig,
  GetProviders,
  TestConnection,
} from "../../../wailsjs/go/main/App";
import { useAppStore } from "../../store/appStore";
import type { ProviderConfig, APIProvider } from "../../types";

export function ApiKeysTab() {
  const { config, patchConfig } = useAppStore();
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
  const [newModelInput, setNewModelInput] = useState<Record<string, string>>(
    {},
  );
  const inputRefs = useRef<Record<string, HTMLInputElement | null>>({});

  useEffect(() => {
    loadConfig();
  }, []);

  const loadConfig = async () => {
    try {
      const cfg = await GetConfig();
      const normalizedProviders: Record<string, ProviderConfig> = {};
      for (const [id, p] of Object.entries(cfg.providers || {})) {
        normalizedProviders[id] = {
          apiKey: p.apiKey || "",
          model: p.model || (p as any).defaultModel || "",
          models: (p as any).models && (p as any).models.length > 0 ? (p as any).models : p.model ? [p.model] : [],
          defaultModel: (p as any).defaultModel || p.model || "",
          enabled: p.enabled ?? false,
        };
      }
      setProviders(normalizedProviders);
      setActiveProvider(cfg.activeProvider || "kilo");
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

  const saveProvider = async (id: string, updated: ProviderConfig) => {
    const next = { ...providers, [id]: updated };
    setProviders(next);
    await UpdateConfig({ providers: next });
    if (config) {
      patchConfig({ providers: next });
    }
  };

  const handleApiKeyChange = async (id: string, apiKey: string) => {
    await saveProvider(id, { ...providers[id], apiKey });
  };

  const handleToggleEnabled = async (id: string) => {
    await saveProvider(id, { ...providers[id], enabled: !providers[id].enabled });
  };

  const handleSetActive = async (id: string) => {
    setActiveProvider(id);
    await UpdateConfig({ activeProvider: id });
    if (config) {
      patchConfig({ activeProvider: id });
    }
  };

  const handleAddModel = async (id: string) => {
    const model = newModelInput[id]?.trim();
    if (!model) return;
    const prov = providers[id];
    if (prov.models.includes(model)) {
      setNewModelInput((prev) => ({ ...prev, [id]: "" }));
      return;
    }
    const updated = {
      ...prov,
      models: [...prov.models, model],
      defaultModel: prov.defaultModel || model,
    };
    setNewModelInput((prev) => ({ ...prev, [id]: "" }));
    await saveProvider(id, updated);
  };

  const handleRemoveModel = async (id: string, modelToRemove: string) => {
    const prov = providers[id];
    const updatedModels = prov.models.filter((m) => m !== modelToRemove);
    let updatedDefault = prov.defaultModel;
    if (prov.defaultModel === modelToRemove) {
      updatedDefault = updatedModels.length > 0 ? updatedModels[0] : "";
    }
    await saveProvider(id, {
      ...prov,
      models: updatedModels,
      defaultModel: updatedDefault,
    });
  };

  const handleSetDefaultModel = async (id: string, model: string) => {
    await saveProvider(id, { ...providers[id], defaultModel: model });
  };

  const handleNewModelKeyDown = (id: string, e: React.KeyboardEvent) => {
    if (e.key === "Enter") {
      e.preventDefault();
      handleAddModel(id);
    }
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
    const modelToTest = config.defaultModel || config.model;
    if (!modelToTest?.trim()) {
      setTestResults((prev) => ({
        ...prev,
        [id]: { success: false, message: "No model selected" },
      }));
      return;
    }
    setTesting(id);
    try {
      const result = await TestConnection(id, config.apiKey, modelToTest);
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
            models: [provider.defaultModel],
            defaultModel: provider.defaultModel,
            enabled: false,
          };
          const isTested = testResults[id];
          const modelList = config.models.length > 0 ? config.models : [config.model || provider.defaultModel].filter(Boolean);

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

                <div className="flex flex-wrap gap-1.5">
                  {modelList.map((model) => (
                    <button
                      key={model}
                      onClick={() => handleSetDefaultModel(id, model)}
                      className={`inline-flex items-center gap-1 px-2 py-0.5 rounded-md text-xs font-mono transition-colors ${
                        config.defaultModel === model
                          ? "bg-foreground text-background"
                          : "bg-foreground/5 text-muted-foreground hover:text-foreground border border-border"
                      }`}
                      title={config.defaultModel === model ? "Default model (click to change)" : "Click to set as default"}
                    >
                      {model}
                      <span
                        role="button"
                        onClick={(e) => {
                          e.stopPropagation();
                          handleRemoveModel(id, model);
                        }}
                        className="ml-0.5 opacity-60 hover:opacity-100"
                      >
                        ×
                      </span>
                    </button>
                  ))}
                </div>

                <div className="flex gap-2">
                  <input
                    ref={(el) => { inputRefs.current[id] = el; }}
                    type="text"
                    value={newModelInput[id] || ""}
                    onChange={(e) =>
                      setNewModelInput((prev) => ({ ...prev, [id]: e.target.value }))
                    }
                    onKeyDown={(e) => handleNewModelKeyDown(id, e)}
                    placeholder="Add model ID..."
                    className="flex-1 h-8 px-3 rounded-md border border-border bg-transparent text-sm placeholder:text-muted-foreground/50 focus:outline-none focus:ring-1 focus:ring-foreground/20 font-mono"
                  />
                  <button
                    onClick={() => handleAddModel(id)}
                    className="h-8 px-3 rounded-md text-xs border border-border hover:bg-foreground/5 transition-colors"
                  >
                    Add
                  </button>
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
