import { useState, useEffect, useCallback, useRef } from 'react'
import { GetConfig, GetGenreList, ListDirectory, SelectDirectory, TranslateFiles, CancelTask, StopAll, ProcessDroppedFiles } from '../../../wailsjs/go/main/App'
import { useAppStore } from '../../store/appStore'
import type { FileInfo, GenrePrompt } from '../../types'
import { StreamingPreview } from '../StreamingPreview'

const VALID_EXTENSIONS = ['.txt', '.md']

export function TranslatePage() {
  const { config, selectedGenre, setSelectedGenre, tasks, clearTasks } = useAppStore()
  const [genres, setGenres] = useState<GenrePrompt[]>([])
  const [selectedDir, setSelectedDir] = useState('')
  const [files, setFiles] = useState<FileInfo[]>([])
  const [selectedFiles, setSelectedFiles] = useState<Set<string>>(new Set())
  const [isTranslating, setIsTranslating] = useState(false)
  const [apiError, setApiError] = useState('')
  const [hoveredTaskId, setHoveredTaskId] = useState<string | null>(null)
  const [isDragging, setIsDragging] = useState(false)
  const dragCounter = useRef(0)

  useEffect(() => {
    GetGenreList().then(setGenres).catch(() => setGenres([]))
    if (!selectedGenre) {
      GetGenreList().then((list) => {
        if (list?.length) setSelectedGenre(list[0].id)
      })
    }
  }, [])

  const taskList = Object.values(tasks)
  const activeTasks = taskList.filter(t => t.status === 'translating' || t.status === 'reviewing' || t.status === 'retrying')
  const doneTasks = taskList.filter(t => t.status === 'done' || t.status === 'cached')
  const errorTasks = taskList.filter(t => t.status === 'error')
  const totalSelected = selectedFiles.size
  const overallProgress = totalSelected > 0
    ? taskList.reduce((sum, t) => sum + t.progress, 0) / totalSelected
    : 0

  const handleCancelTask = async (taskId: string) => {
    try {
      await CancelTask(taskId)
    } catch (e) {
      console.error(e)
    }
  }

  const handleStopAll = async () => {
    try {
      await StopAll()
    } catch (e) {
      console.error(e)
    }
  }

  const checkApiConfig = (): boolean => {
    if (!config) {
      setApiError('Config not loaded. Please wait or reload the app.')
      return false
    }
    const provider = config.providers?.[config.activeProvider]
    if (!provider) {
      setApiError(`No provider configured for "${config.activeProvider}". Go to Settings → API Keys.`)
      return false
    }
    if (!provider.enabled) {
      setApiError(`Provider "${config.activeProvider}" is disabled. Enable it in Settings.`)
      return false
    }
    if (!provider.apiKey?.trim()) {
      setApiError(`API key is empty for "${config.activeProvider}". Add it in Settings.`)
      return false
    }
    if (!provider.model?.trim()) {
      setApiError(`Model is empty for "${config.activeProvider}". Set it in Settings.`)
      return false
    }
    setApiError('')
    return true
  }

  const handleSelectDir = async () => {
    try {
      const dir = await SelectDirectory()
      if (dir) {
        setSelectedDir(dir)
        loadFiles(dir)
      }
    } catch (e) {
      console.error(e)
    }
  }

  const loadFiles = async (dir: string) => {
    try {
      const allFiles = await ListDirectory(dir)
      const validFiles = (allFiles || []).filter(
        (f: FileInfo) => !f.isDir && VALID_EXTENSIONS.some(ext => f.name.toLowerCase().endsWith(ext))
      )
      validFiles.sort((a: FileInfo, b: FileInfo) => a.name.localeCompare(b.name, undefined, { numeric: true }))
      setFiles(validFiles)
      setSelectedFiles(new Set(validFiles.map((f: FileInfo) => f.path)))
    } catch (e) {
      console.error(e)
      setFiles([])
    }
  }

  const toggleFile = (path: string) => {
    setSelectedFiles(prev => {
      const next = new Set(prev)
      if (next.has(path)) next.delete(path)
      else next.add(path)
      return next
    })
  }

  const selectAll = () => setSelectedFiles(new Set(files.map(f => f.path)))
  const deselectAll = () => setSelectedFiles(new Set())

  const readDirectoryEntry = async (dirEntry: any): Promise<{ name: string; content: string }[]> => {
    const reader = dirEntry.createReader()
    const entries: any[] = await new Promise((resolve) => reader.readEntries(resolve))
    const results: { name: string; content: string }[] = []

    for (const entry of entries) {
      if (entry.isFile) {
        const file: File = await new Promise((resolve) => entry.file(resolve))
        if (VALID_EXTENSIONS.some(ext => file.name.toLowerCase().endsWith(ext))) {
          const content = await file.text()
          results.push({ name: file.name, content })
        }
      } else if (entry.isDirectory) {
        const subResults = await readDirectoryEntry(entry)
        results.push(...subResults)
      }
    }

    return results
  }

  const handleDrop = useCallback(async (e: React.DragEvent) => {
    e.preventDefault()
    e.stopPropagation()
    setIsDragging(false)
    dragCounter.current = 0

    const items = e.dataTransfer.items
    if (!items) return

    const droppedFiles: { name: string; content: string }[] = []

    for (let i = 0; i < items.length; i++) {
      const item = items[i]
      const entry = (item as any).webkitGetAsEntry?.()
      if (!entry) continue

      if (entry.isFile) {
        const file: File = await new Promise((resolve) => entry.file(resolve))
        if (VALID_EXTENSIONS.some(ext => file.name.toLowerCase().endsWith(ext))) {
          const content = await file.text()
          droppedFiles.push({ name: file.name, content })
        }
      } else if (entry.isDirectory) {
        const subFiles = await readDirectoryEntry(entry)
        droppedFiles.push(...subFiles)
      }
    }

    if (droppedFiles.length === 0) return

    try {
      const tempDir = await ProcessDroppedFiles(droppedFiles)
      setSelectedDir(tempDir)
      loadFiles(tempDir)
    } catch (e) {
      console.error(e)
    }
  }, [])

  const handleDragOver = useCallback((e: React.DragEvent) => {
    e.preventDefault()
    e.stopPropagation()
    if (!isDragging) setIsDragging(true)
  }, [isDragging])

  const handleDragEnter = useCallback((e: React.DragEvent) => {
    e.preventDefault()
    e.stopPropagation()
    dragCounter.current++
    if (dragCounter.current === 1) setIsDragging(true)
  }, [])

  const handleDragLeave = useCallback((e: React.DragEvent) => {
    e.preventDefault()
    e.stopPropagation()
    dragCounter.current--
    if (dragCounter.current === 0) setIsDragging(false)
  }, [])

  const handleTranslate = async () => {
    const filesToTranslate = files.filter(f => selectedFiles.has(f.path))
    if (filesToTranslate.length === 0) return

    if (!checkApiConfig()) return

    clearTasks()
    setIsTranslating(true)
    setApiError('')
    const paths = filesToTranslate.map(f => f.path)
    try {
      await TranslateFiles(paths, selectedGenre)
    } catch (e: any) {
      setApiError(e.message || String(e))
    } finally {
      setIsTranslating(false)
    }
  }

  return (
    <div
      className="flex-1 flex flex-col overflow-hidden"
      onDragEnter={handleDragEnter}
      onDragOver={handleDragOver}
      onDragLeave={handleDragLeave}
      onDrop={handleDrop}
    >
      <div className="flex-1 overflow-y-auto p-8">
        <div className="max-w-2xl mx-auto space-y-6">
          <div className="text-center space-y-2 mb-8">
            <h1 className="text-2xl font-light tracking-tight">Translate</h1>
            <p className="text-sm text-muted-foreground">Select a directory and choose files to translate</p>
          </div>

          {isDragging && (
            <div className="fixed inset-0 z-50 flex items-center justify-center bg-background/80 backdrop-blur-sm">
              <div className="border-2 border-dashed border-foreground/30 rounded-xl p-16 text-center">
                <svg xmlns="http://www.w3.org/2000/svg" width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round" className="mx-auto mb-4 text-muted-foreground"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><polyline points="17 8 12 3 7 8"/><line x1="12" x2="12" y1="3" y2="15"/></svg>
                <p className="text-lg font-light">Drop folder or files here</p>
                <p className="text-xs text-muted-foreground mt-1">.txt and .md files will be loaded</p>
              </div>
            </div>
          )}

          {apiError && (
            <div className="p-3 rounded-md border border-red-500/30 bg-red-500/5 text-sm text-red-400">
              {apiError}
            </div>
          )}

          <div className="space-y-4">
            <button
              onClick={handleSelectDir}
              className="w-full h-10 rounded-md border border-border text-sm hover:bg-foreground/5 transition-colors text-left px-4 truncate"
            >
              {selectedDir || 'Select directory...'}
            </button>

            {selectedDir && (
              <div className="space-y-2">
                <div className="flex items-center justify-between">
                  <span className="text-xs text-muted-foreground">
                    {files.length} files · {selectedFiles.size} selected
                  </span>
                  <div className="flex gap-2">
                    <button onClick={selectAll} className="text-xs text-muted-foreground hover:text-foreground">
                      Select all
                    </button>
                    <button onClick={deselectAll} className="text-xs text-muted-foreground hover:text-foreground">
                      Deselect all
                    </button>
                  </div>
                </div>

                <div className="space-y-0.5 max-h-48 overflow-y-auto">
                  {files.map((file) => {
                    const taskForFile = taskList.find(t => t.fileName === file.name)
                    const isSelected = selectedFiles.has(file.path)
                    const taskStatus = taskForFile?.status

                    return (
                      <label
                        key={file.path}
                        className={`flex items-center gap-3 px-3 py-2 rounded-md cursor-pointer transition-colors ${
                          isSelected ? 'bg-foreground/5' : 'hover:bg-foreground/5'
                        }`}
                      >
                        <input
                          type="checkbox"
                          checked={isSelected}
                          onChange={() => toggleFile(file.path)}
                          disabled={isTranslating}
                          className="w-3.5 h-3.5 rounded border-border accent-foreground disabled:opacity-50"
                        />
                        <span className="text-sm flex-1 truncate">{file.name}</span>
                        {taskStatus && (
                          <StatusBadge status={taskStatus} error={taskForFile?.error} />
                        )}
                        <span className="text-xs text-muted-foreground w-12 text-right">
                          {(file.size / 1024).toFixed(0)}KB
                        </span>
                      </label>
                    )
                  })}
                </div>
              </div>
            )}

            <div className="flex gap-2">
              {genres.map((g) => (
                <button
                  key={g.id}
                  onClick={() => setSelectedGenre(g.id)}
                  disabled={isTranslating}
                  className={`px-3 h-8 rounded-md text-xs transition-colors ${
                    selectedGenre === g.id
                      ? 'bg-foreground text-background'
                      : 'border border-border text-muted-foreground hover:text-foreground'
                  } disabled:opacity-50`}
                >
                  {g.name}
                </button>
              ))}
            </div>

            <button
              onClick={handleTranslate}
              disabled={isTranslating || selectedFiles.size === 0}
              className="w-full h-10 rounded-md bg-foreground text-background text-sm font-medium hover:opacity-90 transition-opacity disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {isTranslating ? 'Translating...' : `Translate${selectedFiles.size > 1 ? ` ${selectedFiles.size} files` : ''}`}
            </button>
          </div>

          {taskList.length > 0 && (
            <div className="space-y-3 pt-4 border-t border-border">
              <div className="flex justify-between items-center text-xs text-muted-foreground">
                <span>Overall</span>
                <div className="flex items-center gap-3">
                  <span>{Math.round(overallProgress)}% · {doneTasks.length}/{taskList.length} done</span>
                  {activeTasks.length > 0 && (
                    <button
                      onClick={handleStopAll}
                      className="text-red-400 hover:text-red-300 transition-colors"
                    >
                      Stop All
                    </button>
                  )}
                </div>
              </div>
              <div className="h-1 bg-muted rounded-full overflow-hidden">
                <div
                  className="h-full bg-foreground transition-all duration-500"
                  style={{ width: `${overallProgress}%` }}
                />
              </div>

              <div className="space-y-2 mt-2">
                {taskList.map((task) => {
                  const isActive = task.status === 'translating' || task.status === 'reviewing' || task.status === 'retrying'
                  const tokenText = task.tokenUsage && task.tokenUsage.totalTokens > 0
                    ? ` · ${formatTokens(task.tokenUsage.totalTokens)}`
                    : ''

                  return (
                    <div
                      key={task.taskId}
                      className="group relative flex items-center gap-3"
                      onMouseEnter={() => setHoveredTaskId(task.taskId)}
                      onMouseLeave={() => setHoveredTaskId(null)}
                    >
                      <span className="text-xs w-28 truncate text-muted-foreground">{task.fileName}</span>
                      <div className="flex-1 h-0.5 bg-muted rounded-full overflow-hidden">
                        <div
                          className={`h-full transition-all duration-300 ${
                            task.status === 'error' ? 'bg-red-500' :
                            task.status === 'cancelled' ? 'bg-gray-500' :
                            task.status === 'retrying' ? 'bg-orange-500' :
                            task.status === 'cached' ? 'bg-purple-500' :
                            task.status === 'done' ? 'bg-green-500' :
                            task.status === 'reviewing' ? 'bg-yellow-500' :
                            'bg-foreground'
                          }`}
                          style={{ width: `${task.progress}%` }}
                        />
                      </div>
                      <span className="text-xs w-24 text-right text-muted-foreground">
                        {Math.round(task.progress)}%{tokenText}
                      </span>
                      <StatusBadge status={task.status} error={task.error} />
                      {isActive && (
                        <button
                          onClick={() => handleCancelTask(task.taskId)}
                          className="opacity-0 group-hover:opacity-100 text-xs text-red-400 hover:text-red-300 transition-opacity"
                        >
                          Cancel
                        </button>
                      )}
                      {hoveredTaskId === task.taskId && (
                        <StreamingPreview taskId={task.taskId} />
                      )}
                    </div>
                  )
                })}
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  )
}

function formatTokens(n: number): string {
  if (n >= 1000) return `${(n / 1000).toFixed(1)}k tokens`
  return `${n} tokens`
}

function StatusBadge({ status, error }: { status: string; error?: string }) {
  const styles: Record<string, string> = {
    waiting: 'text-muted-foreground',
    translating: 'text-blue-400',
    reviewing: 'text-yellow-400',
    done: 'text-green-400',
    error: 'text-red-400',
    cancelled: 'text-gray-400',
    retrying: 'text-orange-400',
    cached: 'text-purple-400',
  }

  const labels: Record<string, string> = {
    waiting: 'waiting',
    translating: 'translating',
    reviewing: 'reviewing',
    done: 'done',
    error: 'error',
    cancelled: 'cancelled',
    retrying: 'retrying',
    cached: 'cached',
  }

  return (
    <span
      className={`text-xs ${styles[status] || 'text-muted-foreground'} ${error ? 'cursor-help' : ''}`}
      title={error || ''}
    >
      {labels[status] || status}
    </span>
  )
}
