import { useEffect, useState } from "react";
import { GetConfig } from "../wailsjs/go/main/App";
import {
  Quit,
  WindowMinimise,
  WindowToggleMaximise,
} from "../wailsjs/runtime/runtime";
import { useAppStore } from "./store/appStore";
import { useWailsEvents } from "./hooks/useWailsEvents";
import { Sidebar } from "./components/Sidebar";
import { TranslatePage } from "./components/pages/Translate";
import { PromptsPage } from "./components/pages/Prompts";
import { GlossaryPage } from "./components/pages/Glossary";
import { SettingsPopup } from "./components/SettingsPopup";

function TitleBar({ onSettings }: { onSettings: () => void }) {
  const noDrag = { WebkitAppRegion: "no-drag" } as React.CSSProperties;

  return (
    <div
      className="h-8 flex items-center px-3 select-none flex-shrink-0"
      style={{ WebkitAppRegion: "drag" } as React.CSSProperties}
    >
      <div className="flex items-center gap-2">
        <div
          className="w-3 h-3 rounded-full bg-[#ff5f57] cursor-pointer hover:brightness-125 transition-all"
          style={noDrag}
          onClick={() => Quit()}
          title="Close"
        />
        <div
          className="w-3 h-3 rounded-full bg-[#febc2e] cursor-pointer hover:brightness-125 transition-all"
          style={noDrag}
          onClick={() => WindowMinimise()}
          title="Minimize"
        />
        <div
          className="w-3 h-3 rounded-full bg-[#28c840] cursor-pointer hover:brightness-125 transition-all"
          style={noDrag}
          onClick={() => WindowToggleMaximise()}
          title="Maximize"
        />
      </div>
      <div className="flex-1 text-center text-xs text-muted-foreground">
        novella
      </div>
      <div
        className="w-8 h-8 flex items-center justify-center text-muted-foreground hover:text-foreground transition-colors cursor-pointer"
        style={{ WebkitAppRegion: "no-drag" } as React.CSSProperties}
        onClick={onSettings}
        title="Settings"
      >
        <svg
          xmlns="http://www.w3.org/2000/svg"
          width="14"
          height="14"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          strokeWidth="2"
          strokeLinecap="round"
          strokeLinejoin="round"
        >
          <path d="M12.22 2h-.44a2 2 0 0 0-2 2v.18a2 2 0 0 1-1 1.73l-.43.25a2 2 0 0 1-2 0l-.15-.08a2 2 0 0 0-2.73.73l-.22.38a2 2 0 0 0 .73 2.73l.15.1a2 2 0 0 1 1 1.72v.51a2 2 0 0 1-1 1.74l-.15.09a2 2 0 0 0-.73 2.73l.22.38a2 2 0 0 0 2.73.73l.15-.08a2 2 0 0 1 2 0l.43.25a2 2 0 0 1 1 1.73V20a2 2 0 0 0 2 2h.44a2 2 0 0 0 2-2v-.18a2 2 0 0 1 1-1.73l.43-.25a2 2 0 0 1 2 0l.15.08a2 2 0 0 0 2.73-.73l.22-.39a2 2 0 0 0-.73-2.73l-.15-.08a2 2 0 0 1-1-1.74v-.5a2 2 0 0 1 1-1.74l.15-.09a2 2 0 0 0 .73-2.73l-.22-.38a2 2 0 0 0-2.73-.73l-.15.08a2 2 0 0 1-2 0l-.43-.25a2 2 0 0 1-1-1.73V4a2 2 0 0 0-2-2z" />
          <circle cx="12" cy="12" r="3" />
        </svg>
      </div>
    </div>
  );
}

export default function App() {
  const { activePage, config, setConfig, setTheme } = useAppStore();
  const [settingsOpen, setSettingsOpen] = useState(false);
  useWailsEvents();

  useEffect(() => {
    GetConfig()
      .then((cfg) => {
        const normalized = {
          ...cfg,
          providers: Object.fromEntries(
            Object.entries(cfg.providers || {}).map(([id, p]) => [
              id,
              {
                ...p,
                models: (p as any).models || (p as any).model ? [(p as any).model] : [],
                defaultModel: (p as any).defaultModel || (p as any).model || "",
              },
            ])
          ),
        };
        setConfig(normalized as any);
        if (cfg.theme) {
          setTheme(cfg.theme);
          document.documentElement.classList.toggle(
            "dark",
            cfg.theme === "dark",
          );
        }
      })
      .catch(console.error);
  }, []);

  const renderPage = () => {
    switch (activePage) {
      case "prompts":
        return <PromptsPage />;
      case "glossary":
        return <GlossaryPage />;
      default:
        return <TranslatePage />;
    }
  };

  return (
    <div className="h-screen flex flex-col bg-background text-foreground overflow-hidden">
      <TitleBar onSettings={() => setSettingsOpen(true)} />
      <div className="flex-1 flex overflow-hidden">
        <Sidebar />
        {renderPage()}
      </div>
      <SettingsPopup
        open={settingsOpen}
        onClose={() => setSettingsOpen(false)}
      />
    </div>
  );
}
