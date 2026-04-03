import { useAppStore } from "../store/appStore";

interface TaskPreviewProps {
  taskId: string;
}

function formatTokens(n: number): string {
  if (n >= 1000) return `${(n / 1000).toFixed(1)}k`;
  return `${n}`;
}

function formatChars(n: number): string {
  if (n >= 1000) return `${(n / 1000).toFixed(1)}k`;
  return `${n}`;
}

export function StreamingPreview({ taskId }: TaskPreviewProps) {
  const task = useAppStore.getState().tasks[taskId];
  const streamContent = useAppStore.getState().streamContent[taskId] || "";

  if (!task) return null;

  const statusColors: Record<string, string> = {
    translating: "text-blue-400",
    reviewing: "text-yellow-400",
    retrying: "text-orange-400",
    done: "text-green-400",
    cached: "text-purple-400",
    error: "text-red-400",
    cancelled: "text-gray-400",
  };

  const statusLabels: Record<string, string> = {
    translating: "Translating",
    reviewing: "Reviewing",
    retrying: "Retrying",
    done: "Done",
    cached: "Cached",
    error: "Error",
    cancelled: "Cancelled",
  };

  return (
    <div className="absolute left-full ml-3 top-0 z-50 w-80 rounded-lg border border-border bg-card shadow-xl p-4 pointer-events-none">
      <div className="flex items-center justify-between mb-3">
        <span className="text-sm font-medium truncate">{task.fileName}</span>
        <span
          className={`text-[10px] uppercase tracking-wider ${statusColors[task.status] || "text-muted-foreground"}`}
        >
          {statusLabels[task.status] || task.status}
        </span>
      </div>

      <div className="space-y-2 mb-3">
        <div className="h-1 bg-muted rounded-full overflow-hidden">
          <div
            className={`h-full transition-all duration-300 ${
              task.status === "error"
                ? "bg-red-500"
                : task.status === "done"
                  ? "bg-green-500"
                  : task.status === "cached"
                    ? "bg-purple-500"
                    : "bg-foreground"
            }`}
            style={{ width: `${task.progress}%` }}
          />
        </div>
        <div className="flex justify-between text-[10px] text-muted-foreground">
          <span>{Math.round(task.progress)}%</span>
          {task.totalChunks > 0 && (
            <span>
              Chunk {task.currentChunk}/{task.totalChunks}
            </span>
          )}
        </div>
      </div>

      <div className="grid grid-cols-2 gap-2 text-[10px] text-muted-foreground mb-3">
        <div>
          <span className="uppercase tracking-wider">Chars</span>
          <p className="text-foreground">
            {formatChars(task.processedChars)} / {formatChars(task.totalChars)}
          </p>
        </div>
        {task.tokenUsage && task.tokenUsage.totalTokens > 0 && (
          <div>
            <span className="uppercase tracking-wider">Tokens</span>
            <p className="text-foreground">
              {formatTokens(task.tokenUsage.totalTokens)}
            </p>
          </div>
        )}
      </div>

      {streamContent && (
        <div>
          <div className="text-[10px] uppercase tracking-wider text-muted-foreground mb-1">
            Latest chunk
          </div>
          <pre className="text-xs text-foreground/80 font-mono whitespace-pre-wrap break-words max-h-32 overflow-y-auto leading-relaxed">
            {streamContent.length > 400
              ? streamContent.slice(0, 400) + "..."
              : streamContent}
          </pre>
        </div>
      )}
    </div>
  );
}
