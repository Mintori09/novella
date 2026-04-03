import { useState, useEffect, useCallback } from 'react'
import { GetGenreList, AddGenrePrompt, UpdateGenrePrompt, RemoveGenrePrompt, GetGenrePrompt } from '../../../wailsjs/go/main/App'
import type { GenrePrompt } from '../../types'

export function PromptsPage() {
  const [genres, setGenres] = useState<GenrePrompt[]>([])
  const [selectedGenre, setSelectedGenre] = useState<GenrePrompt | null>(null)
  const [isEditing, setIsEditing] = useState(false)
  const [editContent, setEditContent] = useState('')
  const [isAdding, setIsAdding] = useState(false)
  const [newId, setNewId] = useState('')
  const [newName, setNewName] = useState('')
  const [newContent, setNewContent] = useState('')
  const [saveMsg, setSaveMsg] = useState('')

  useEffect(() => {
    loadGenres()
  }, [])

  const loadGenres = async () => {
    try {
      const list = await GetGenreList()
      setGenres(list || [])
    } catch (e) {
      console.error(e)
    }
  }

  const handleSelect = async (genre: GenrePrompt) => {
    try {
      const full = await GetGenrePrompt(genre.id)
      setSelectedGenre(full)
      setEditContent(full.content)
      setIsEditing(false)
    } catch (e) {
      console.error(e)
    }
  }

  const handleSaveEdit = async () => {
    if (!selectedGenre) return
    try {
      await UpdateGenrePrompt(selectedGenre.id, editContent)
      setIsEditing(false)
      setSaveMsg('Saved!')
      setTimeout(() => setSaveMsg(''), 2000)
      loadGenres()
      handleSelect(selectedGenre)
    } catch (e: any) {
      setSaveMsg(e.message || 'Save failed')
    }
  }

  const handleAdd = async () => {
    if (!newId.trim() || !newName.trim() || !newContent.trim()) return
    try {
      await AddGenrePrompt(newId.trim().toLowerCase().replace(/\s+/g, '_'), newName.trim(), newContent.trim())
      setIsAdding(false)
      setNewId('')
      setNewName('')
      setNewContent('')
      setSaveMsg('Genre added!')
      setTimeout(() => setSaveMsg(''), 2000)
      loadGenres()
    } catch (e: any) {
      setSaveMsg(e.message || 'Add failed')
    }
  }

  const handleDelete = useCallback(async (id: string) => {
    if (!confirm(`Delete genre "${id}"? This cannot be undone.`)) return
    try {
      await RemoveGenrePrompt(id)
      if (selectedGenre?.id === id) {
        setSelectedGenre(null)
      }
      setSaveMsg('Genre deleted')
      setTimeout(() => setSaveMsg(''), 2000)
      loadGenres()
    } catch (e: any) {
      setSaveMsg(e.message || 'Delete failed')
    }
  }, [selectedGenre])

  const generateId = (name: string) => {
    return name.toLowerCase()
      .replace(/\s+/g, '_')
      .replace(/[^a-z0-9_]/g, '')
  }

  return (
    <div className="flex-1 flex flex-col overflow-hidden">
      <div className="flex-1 overflow-y-auto p-8">
        <div className="max-w-4xl mx-auto space-y-6">
          <div className="text-center space-y-2 mb-8">
            <h1 className="text-2xl font-light tracking-tight">Prompts</h1>
            <p className="text-sm text-muted-foreground">Manage genre prompts for translation</p>
          </div>

          {saveMsg && (
            <div className="p-3 rounded-md border border-green-500/30 bg-green-500/5 text-sm text-green-400 text-center">
              {saveMsg}
            </div>
          )}

          <div className="flex gap-6">
            {/* Genre List */}
            <div className="w-64 flex-shrink-0 space-y-2">
              <div className="flex items-center justify-between mb-2">
                <span className="text-xs text-muted-foreground uppercase tracking-wider">Genres</span>
                <button
                  onClick={() => setIsAdding(true)}
                  className="text-xs text-foreground hover:text-foreground/70 transition-colors"
                  title="Add new genre"
                >
                  + Add
                </button>
              </div>

              <div className="space-y-0.5 max-h-[500px] overflow-y-auto">
                {genres.map((g) => (
                  <div
                    key={g.id}
                    className={`flex items-center gap-2 px-3 py-2 rounded-md cursor-pointer transition-colors ${
                      selectedGenre?.id === g.id ? 'bg-foreground/10' : 'hover:bg-foreground/5'
                    }`}
                    onClick={() => handleSelect(g)}
                  >
                    <span className="text-sm flex-1 truncate">{g.name}</span>
                    {g.builtin && (
                      <span className="text-[10px] text-muted-foreground px-1.5 py-0.5 rounded bg-foreground/5">
                        built-in
                      </span>
                    )}
                    {!g.builtin && (
                      <button
                        onClick={(e) => { e.stopPropagation(); handleDelete(g.id) }}
                        className="text-[10px] text-red-400 hover:text-red-300 transition-colors"
                        title="Delete"
                      >
                        ×
                      </button>
                    )}
                  </div>
                ))}
              </div>
            </div>

            {/* Editor */}
            <div className="flex-1 space-y-4">
              {selectedGenre ? (
                <>
                  <div className="flex items-center justify-between">
                    <div>
                      <h2 className="text-lg font-light">{selectedGenre.name}</h2>
                      <p className="text-xs text-muted-foreground font-mono">{selectedGenre.id}</p>
                    </div>
                    {!isEditing ? (
                      <button
                        onClick={() => setIsEditing(true)}
                        className="h-8 px-3 rounded-md text-xs border border-border hover:bg-foreground/5 transition-colors"
                      >
                        Edit
                      </button>
                    ) : (
                      <div className="flex gap-2">
                        <button
                          onClick={() => { setIsEditing(false); setEditContent(selectedGenre.content) }}
                          className="h-8 px-3 rounded-md text-xs border border-border hover:bg-foreground/5 transition-colors"
                        >
                          Cancel
                        </button>
                        <button
                          onClick={handleSaveEdit}
                          className="h-8 px-3 rounded-md text-xs bg-foreground text-background hover:opacity-90 transition-opacity"
                        >
                          Save
                        </button>
                      </div>
                    )}
                  </div>

                  {isEditing ? (
                    <textarea
                      value={editContent}
                      onChange={(e) => setEditContent(e.target.value)}
                      className="w-full h-[400px] p-4 rounded-md border border-border bg-transparent text-sm font-mono resize-none focus:outline-none focus:ring-1 focus:ring-foreground/20"
                      spellCheck={false}
                    />
                  ) : (
                    <pre className="w-full h-[400px] p-4 rounded-md border border-border bg-foreground/5 text-sm font-mono overflow-y-auto whitespace-pre-wrap">
                      {selectedGenre.content}
                    </pre>
                  )}
                </>
              ) : (
                <div className="flex items-center justify-center h-[400px] text-muted-foreground text-sm">
                  Select a genre to view its prompt
                </div>
              )}
            </div>
          </div>

          {/* Add Genre Modal */}
          {isAdding && (
            <div className="fixed inset-0 z-50 flex items-center justify-center bg-background/80 backdrop-blur-sm">
              <div className="relative w-[600px] bg-card border border-border rounded-xl shadow-2xl p-6 space-y-4">
                <button
                  onClick={() => { setIsAdding(false); setNewId(''); setNewName(''); setNewContent('') }}
                  className="absolute top-3 right-3 w-6 h-6 flex items-center justify-center text-muted-foreground hover:text-foreground transition-colors"
                >
                  <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><path d="M18 6 6 18"/><path d="m6 6 12 12"/></svg>
                </button>

                <h3 className="text-lg font-light">Add New Genre</h3>

                <div className="space-y-3">
                  <div className="space-y-1">
                    <label className="text-xs text-muted-foreground uppercase tracking-wider">Name</label>
                    <input
                      type="text"
                      value={newName}
                      onChange={(e) => {
                        setNewName(e.target.value)
                        setNewId(generateId(e.target.value))
                      }}
                      placeholder="e.g. Kiếm hiệp"
                      className="w-full h-8 px-3 rounded-md border border-border bg-transparent text-sm focus:outline-none focus:ring-1 focus:ring-foreground/20"
                    />
                  </div>

                  <div className="space-y-1">
                    <label className="text-xs text-muted-foreground uppercase tracking-wider">ID</label>
                    <input
                      type="text"
                      value={newId}
                      onChange={(e) => setNewId(e.target.value)}
                      placeholder="e.g. kiem_hiep"
                      className="w-full h-8 px-3 rounded-md border border-border bg-transparent text-sm font-mono focus:outline-none focus:ring-1 focus:ring-foreground/20"
                    />
                  </div>

                  <div className="space-y-1">
                    <label className="text-xs text-muted-foreground uppercase tracking-wider">Prompt Content</label>
                    <textarea
                      value={newContent}
                      onChange={(e) => setNewContent(e.target.value)}
                      placeholder="Write your genre-specific prompt instructions here..."
                      className="w-full h-48 p-3 rounded-md border border-border bg-transparent text-sm font-mono resize-none focus:outline-none focus:ring-1 focus:ring-foreground/20"
                    />
                  </div>

                  <button
                    onClick={handleAdd}
                    disabled={!newId.trim() || !newName.trim() || !newContent.trim()}
                    className="w-full h-10 rounded-md bg-foreground text-background text-sm font-medium hover:opacity-90 transition-opacity disabled:opacity-50 disabled:cursor-not-allowed"
                  >
                    Add Genre
                  </button>
                </div>
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  )
}
