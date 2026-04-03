import { useState } from "react";
import { ApiKeysTab } from "./tabs/ApiKeysTab";
import { SettingsTab } from "./tabs/SettingsTab";

const tabs = [
  { id: "api", label: "API Keys" },
  { id: "settings", label: "Settings" },
];

interface SettingsPopupProps {
  open: boolean;
  onClose: () => void;
}

export function SettingsPopup({ open, onClose }: SettingsPopupProps) {
  const [activeTab, setActiveTab] = useState("api");

  if (!open) return null;

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center">
      <div
        className="absolute inset-0 bg-background/80 backdrop-blur-sm"
        onClick={onClose}
      />
      <div className="relative w-[640px] max-h-[80vh] bg-card border border-border rounded-xl shadow-2xl overflow-hidden flex">
        <div className="w-40 border-r border-border flex flex-col py-4">
          {tabs.map((tab) => (
            <button
              key={tab.id}
              onClick={() => setActiveTab(tab.id)}
              className={`px-4 py-2 text-sm text-left transition-colors ${
                activeTab === tab.id
                  ? "text-foreground bg-foreground/5 border-r-2 border-foreground"
                  : "text-muted-foreground hover:text-foreground hover:bg-foreground/5"
              }`}
            >
              {tab.label}
            </button>
          ))}
        </div>

        <div className="flex-1 overflow-y-auto p-6">
          {activeTab === "api" ? <ApiKeysTab /> : <SettingsTab />}
        </div>

        <button
          onClick={onClose}
          className="absolute top-3 right-3 w-6 h-6 flex items-center justify-center text-muted-foreground hover:text-foreground transition-colors"
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
            <path d="M18 6 6 18" />
            <path d="m6 6 12 12" />
          </svg>
        </button>
      </div>
    </div>
  );
}
