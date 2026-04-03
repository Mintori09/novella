import { useState, useEffect } from 'react'
import { ListGlossaries, GetGlossary, SaveGlossary, AddGlossaryEntry, RemoveGlossaryEntry } from '../../../wailsjs/go/main/App'
import type { Glossary, GlossaryEntry } from '../../types'

export function GlossaryPage() {
  const [glossaries, setGlossaries] = useState<string[]>([])
  const [selected, setSelected] = useState<string>('')
  const [glossary, setGlossary] = useState<Glossary | null>(null)
  const [newEntry, setNewEntry] = useState({ original: '', translated: '', category: 'characters', note: '' })

  useEffect(() => {
    loadGlossaries()
  }, [])

  const loadGlossaries = async () => {
    try {
      const list = await ListGlossaries()
      setGlossaries(list || [])
    } catch (e) {
      console.error(e)
    }
  }

  const loadGlossary = async (id: string) => {
    setSelected(id)
    try {
      const g = await GetGlossary(id)
      setGlossary(g)
    } catch (e) {
      console.error(e)
    }
  }

  const handleAddEntry = async () => {
    if (!selected || !newEntry.original) return
    try {
      await AddGlossaryEntry(selected, newEntry)
      setNewEntry({ original: '', translated: '', category: 'characters', note: '' })
      loadGlossary(selected)
    } catch (e) {
      console.error(e)
    }
  }

  const handleRemoveEntry = async (original: string) => {
    if (!selected) return
    try {
      await RemoveGlossaryEntry(selected, original)
      loadGlossary(selected)
    } catch (e) {
      console.error(e)
    }
  }

  const entries = glossary ? Object.values(glossary.entries || {}) : []

  return (
    <div className="flex-1 overflow-y-auto p-8">
      <div className="max-w-3xl mx-auto space-y-6">
        <div className="space-y-1">
          <h1 className="text-2xl font-light tracking-tight">Glossary</h1>
          <p className="text-sm text-muted-foreground">Manage translation terms per novel</p>
        </div>

        <div className="flex gap-2 flex-wrap">
          {glossaries.map((id) => (
            <button
              key={id}
              onClick={() => loadGlossary(id)}
              className={`px-3 h-8 rounded-md text-xs transition-colors ${
                selected === id
                  ? 'bg-foreground text-background'
                  : 'border border-border text-muted-foreground hover:text-foreground'
              }`}
            >
              {id}
            </button>
          ))}
          {glossaries.length === 0 && (
            <p className="text-sm text-muted-foreground">No glossaries yet. They are auto-extracted from the first 8000 chars during translation.</p>
          )}
        </div>

        {selected && (
          <>
            <div className="flex gap-2">
              <input
                value={newEntry.original}
                onChange={(e) => setNewEntry({ ...newEntry, original: e.target.value })}
                placeholder="Original"
                className="flex-1 h-8 px-3 rounded-md border border-border bg-transparent text-sm placeholder:text-muted-foreground/50 focus:outline-none"
              />
              <input
                value={newEntry.translated}
                onChange={(e) => setNewEntry({ ...newEntry, translated: e.target.value })}
                placeholder="Translation"
                className="flex-1 h-8 px-3 rounded-md border border-border bg-transparent text-sm placeholder:text-muted-foreground/50 focus:outline-none"
              />
              <select
                value={newEntry.category}
                onChange={(e) => setNewEntry({ ...newEntry, category: e.target.value })}
                className="h-8 px-2 rounded-md border border-border bg-transparent text-sm focus:outline-none"
              >
                <option value="characters">Characters</option>
                <option value="locations">Locations</option>
                <option value="realms">Realms</option>
                <option value="techniques">Techniques</option>
                <option value="terms">Terms</option>
              </select>
              <button
                onClick={handleAddEntry}
                className="h-8 px-3 rounded-md bg-foreground text-background text-xs hover:opacity-90 transition-opacity"
              >
                Add
              </button>
            </div>

            {entries.length > 0 && (
              <div className="space-y-1">
                {entries.map((entry) => (
                  <div key={entry.original} className="flex items-center justify-between py-2 border-b border-border/50">
                    <div className="flex items-center gap-4">
                      <span className="text-sm font-mono">{entry.original}</span>
                      <span className="text-muted-foreground">→</span>
                      <span className="text-sm">{entry.translated}</span>
                      <span className="text-xs text-muted-foreground px-2 py-0.5 rounded bg-muted">{entry.category}</span>
                    </div>
                    <button
                      onClick={() => handleRemoveEntry(entry.original)}
                      className="text-muted-foreground hover:text-foreground text-xs"
                    >
                      ×
                    </button>
                  </div>
                ))}
              </div>
            )}
          </>
        )}
      </div>
    </div>
  )
}
