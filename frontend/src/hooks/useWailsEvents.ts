import { useEffect } from "react";
import { useAppStore } from "../store/appStore";
import type { TaskProgress, ChunkResult } from "../types";

declare global {
  interface Window {
    runtime: any;
  }
}

interface FileTask {
  taskId: string;
  fileName: string;
  status:
    | "waiting"
    | "translating"
    | "reviewing"
    | "done"
    | "error"
    | "cancelled"
    | "retrying"
    | "cached";
  progress: number;
  processedChars: number;
  totalChars: number;
  isReviewing: boolean;
  currentChunk: number;
  totalChunks: number;
  error?: string;
  tokenUsage?: {
    inputTokens: number;
    outputTokens: number;
    totalTokens: number;
  };
}

export function useWailsEvents() {
  const { setTask, setStreamContent } = useAppStore.getState();

  useEffect(() => {
    const { runtime } = window;

    if (!runtime?.EventsOn) return;

    runtime.EventsOn(
      "task:started",
      (data: { taskId: string; fileName: string; genre: string }) => {
        setTask(data.taskId, {
          taskId: data.taskId,
          fileName: data.fileName,
          status: "translating",
          progress: 0,
          processedChars: 0,
          totalChars: 0,
          isReviewing: false,
          currentChunk: 0,
          totalChunks: 0,
        });
        setStreamContent(data.taskId, "");
      },
    );

    runtime.EventsOn("task:progress", (data: TaskProgress) => {
      let status: FileTask["status"] = "translating";
      if (data.status === "completed") status = "done";
      else if (data.status === "error") status = "error";
      else if (data.status === "cancelled") status = "cancelled";
      else if (data.status === "retrying") status = "retrying";
      else if (data.status === "cached") status = "cached";
      else if (data.isReviewing) status = "reviewing";

      setTask(data.taskId, {
        taskId: data.taskId,
        fileName: data.fileName,
        status,
        progress: data.progress,
        processedChars: data.processedChars,
        totalChars: data.totalChars,
        isReviewing: data.isReviewing,
        currentChunk: data.currentChunk,
        totalChunks: data.totalChunks,
        error: data.error || undefined,
        tokenUsage:
          data.tokenUsage.totalTokens > 0 ? data.tokenUsage : undefined,
      });
    });

    runtime.EventsOn("chunk:complete", (_result: ChunkResult) => {
      // Handled via task:progress
    });

    return () => {
      runtime.EventsOff("task:started");
      runtime.EventsOff("task:progress");
      runtime.EventsOff("chunk:complete");
    };
  }, []);
}
