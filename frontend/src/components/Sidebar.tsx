import { useAppStore } from "../store/appStore";

const pages = [
  { id: "translate", label: "Translate", icon: "BookOpen" },
  { id: "prompts", label: "Prompts", icon: "FileText" },
  { id: "glossary", label: "Glossary", icon: "BookMarked" },
];

export function Sidebar() {
  const { activePage, setActivePage } = useAppStore();

  return (
    <aside className="w-12 flex flex-col items-center py-4 gap-1 border-r border-border">
      <div className="mb-6">
        <div className="w-8 h-8 rounded-lg bg-foreground/10 flex items-center justify-center">
          <span className="text-xs font-semibold">N</span>
        </div>
      </div>
      {pages.map((page) => (
        <button
          key={page.id}
          onClick={() => setActivePage(page.id)}
          className={`w-8 h-8 rounded-md flex items-center justify-center transition-colors ${
            activePage === page.id
              ? "bg-foreground/10 text-foreground"
              : "text-muted-foreground hover:text-foreground hover:bg-foreground/5"
          }`}
          title={page.label}
        >
          <SidebarIcon name={page.icon} />
        </button>
      ))}
    </aside>
  );
}

function SidebarIcon({ name }: { name: string }) {
  switch (name) {
    case "BookOpen":
      return (
        <svg
          xmlns="http://www.w3.org/2000/svg"
          width="16"
          height="16"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          strokeWidth="2"
          strokeLinecap="round"
          strokeLinejoin="round"
        >
          <path d="M2 3h6a4 4 0 0 1 4 4v14a3 3 0 0 0-3-3H2z" />
          <path d="M22 3h-6a4 4 0 0 0-4 4v14a3 3 0 0 1 3-3h7z" />
        </svg>
      );
    case "FileText":
      return (
        <svg
          xmlns="http://www.w3.org/2000/svg"
          width="16"
          height="16"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          strokeWidth="2"
          strokeLinecap="round"
          strokeLinejoin="round"
        >
          <path d="M15 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V7Z" />
          <path d="M14 2v4a2 2 0 0 0 2 2h4" />
          <path d="M10 9H8" />
          <path d="M16 13H8" />
          <path d="M16 17H8" />
        </svg>
      );
    case "BookMarked":
      return (
        <svg
          xmlns="http://www.w3.org/2000/svg"
          width="16"
          height="16"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          strokeWidth="2"
          strokeLinecap="round"
          strokeLinejoin="round"
        >
          <path d="M10 2v8l3-3 3 3V2" />
          <path d="M4 19.5v-15A2.5 2.5 0 0 1 6.5 2H19a1 1 0 0 1 1 1v18a1 1 0 0 1-1 1H6.5a1 1 0 0 1 0-5H20" />
        </svg>
      );
    default:
      return null;
  }
}
