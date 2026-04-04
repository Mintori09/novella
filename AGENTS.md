# AGENTS.md — Novella

AI-powered literary translation desktop app (Wails v2 + Go backend + React/TypeScript frontend). Translates Chinese/Japanese novels to Vietnamese.

## Project Structure

```
novella/
├── main.go                  # Wails entry point, binds Go methods to frontend
├── wails.json               # Wails config
├── go.mod                   # Go 1.23 module
├── backend/                 # Go packages
│   ├── cache/               # Translation cache manager
│   ├── config/              # App config (JSON file)
│   ├── engine/              # Core translation engine (chunking, retry, review)
│   ├── exporter/            # Output formats (TXT, EPUB)
│   ├── glossary/            # Per-novel glossary management
│   ├── models/              # Shared data types/structs
│   ├── prompts/             # Genre-based prompt manager
│   └── provider/            # AI API providers (OpenAI, OpenRouter, Kilo)
├── frontend/                # React + TypeScript frontend
│   ├── src/
│   │   ├── App.tsx          # Main app with title bar + routing
│   │   ├── main.tsx         # React entry point
│   │   ├── types/index.ts   # TS interfaces (mirrors Go models)
│   │   ├── store/appStore.ts# Zustand state management
│   │   ├── hooks/           # Custom hooks (e.g. useWailsEvents)
│   │   └── components/      # UI components (pages, tabs, popups)
│   ├── package.json
│   └── vite.config.ts
└── prompts/                 # Built-in prompt templates (embedded in Go binary)
```

## Commands

### Go Backend
```bash
wails dev          # Live dev mode (hot reload frontend + Go bridge)
wails build        # Build production redistributable package
go test ./backend/...          # Run all Go tests
go test -run TestName ./backend/engine/  # Run a single Go test
```

### Frontend (run from `frontend/`)
```bash
npm run dev        # Vite dev server (port 5173, strict port)
npm run build      # Type-check (tsc) then build (vite build)
npm run preview    # Preview production build
```

### Testing
**No tests currently exist.** No test framework is installed for frontend. If adding tests:
- **Go:** Standard `go test` — place `*_test.go` files alongside source
- **Frontend:** Would need to add Vitest/Jest + React Testing Library

## Code Style

### Go Backend
- **Formatting:** `gofmt` (tabs for indentation)
- **Imports:** Standard library first, then local packages, then third-party. Grouped in a single `import()` block with blank lines between groups. Local packages use full module path: `"novella/backend/models"`. Use aliases for disambiguation: `promptsmgr "novella/backend/prompts"`
- **Naming:** Types/PascalCase (`AppConfig`), Interfaces use `-er` suffix (`Provider`), exported functions PascalCase, unexported camelCase, variables camelCase, constants UPPER_SNAKE_CASE
- **Struct fields:** camelCase with JSON tags: `json:"apiKey"`
- **Error handling:** Return errors as last value. Use `fmt.Errorf` for context. Do NOT silently swallow errors with `_` unless intentional
- **Concurrency:** Use `sync.Mutex`/`sync.RWMutex` for shared state, `context.Context` for cancellation, goroutines for async work
- **No `any` types** — use specific types or interfaces

### TypeScript Frontend
- **Formatting:** Consistent quote style (prefer single quotes), semicolons, trailing commas
- **Imports:** External libraries first, then local. Use `import type { ... }` for type-only imports
- **Naming:** Components PascalCase (`TranslatePage`), hooks camelCase with `use` prefix (`useAppStore`), types/interfaces PascalCase, variables/functions camelCase, constants UPPER_SNAKE_CASE
- **Components:** Function components with named exports. Main `App` uses default export
- **Styling:** Tailwind CSS exclusively with `darkMode: ["class"]`. shadcn/ui-style CSS variables for theming (HSL). No CSS modules or styled-components
- **State:** Zustand for global state (`appStore.ts`), `useState` for component-level
- **Error handling:** `try/catch` with `console.error(e)`. Display API errors in UI via state
- **TypeScript config:** `strict: false` — but prefer explicit types. Avoid `any`; use proper interfaces from `types/index.ts`

### General
- **No ESLint/Prettier config exists** — follow existing file conventions
- **No `.cursorrules`, `.cursor/rules/`, or `.github/copilot-instructions.md` exist**
- Keep components small; define helper components in the same file
- Use inline SVGs for icons (existing pattern, despite `lucide-react` being a dependency)
- Mirror Go model types in `frontend/src/types/index.ts` when adding/changing data structures
