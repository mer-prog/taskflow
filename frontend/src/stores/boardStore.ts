"use client";

import { create } from "zustand";
import { apiFetch } from "@/lib/api";
import { useAuthStore } from "./authStore";

export interface Label {
  id: string;
  name: string;
  color: string;
}

export interface Assignee {
  id: string;
  display_name: string;
  avatar_url?: string;
}

export interface Task {
  id: string;
  title: string;
  position: number;
  priority: string;
  due_date: string | null;
  assignee: Assignee | null;
  labels: Label[];
}

export interface Column {
  id: string;
  name: string;
  position: number;
  color: string;
  wip_limit: number | null;
  tasks: Task[];
}

export interface BoardDetail {
  id: string;
  project_id: string;
  name: string;
  columns: Column[];
  created_at: string;
  updated_at: string;
}

export interface TaskDetail {
  id: string;
  column_id: string;
  title: string;
  description?: string;
  position: number;
  priority: string;
  assignee: Assignee | null;
  due_date: string | null;
  labels: Label[];
  created_at: string;
  updated_at: string;
}

export interface Comment {
  id: string;
  user_id: string;
  display_name: string;
  content: string;
  created_at: string;
  updated_at: string;
}

interface BoardState {
  boardId: string | null;
  boardName: string;
  columns: Column[];
  activeTask: Task | null;
  isLoading: boolean;
  selectedTask: TaskDetail | null;
  comments: Comment[];

  fetchBoard: (boardId: string) => Promise<void>;
  setActiveTask: (task: Task | null) => void;
  moveTask: (
    taskId: string,
    fromColumnId: string,
    toColumnId: string,
    newPosition: number
  ) => Promise<void>;
  createTask: (columnId: string, title: string) => Promise<void>;
  updateTask: (taskId: string, data: Record<string, unknown>) => Promise<void>;
  deleteTask: (taskId: string, columnId: string) => Promise<void>;
  createColumn: (boardId: string, name: string) => Promise<void>;
  updateColumn: (columnId: string, data: Record<string, unknown>) => Promise<void>;
  deleteColumn: (columnId: string) => Promise<void>;
  fetchTaskDetail: (taskId: string) => Promise<void>;
  fetchComments: (taskId: string) => Promise<void>;
  addComment: (taskId: string, content: string) => Promise<void>;
  clearSelectedTask: () => void;
  handleWSMessage: (message: { type: string; payload: unknown; user_id: string }) => void;
}

function getTenantId() {
  return useAuthStore.getState().tenantId || "";
}

export const useBoardStore = create<BoardState>((set, get) => ({
  boardId: null,
  boardName: "",
  columns: [],
  activeTask: null,
  isLoading: false,
  selectedTask: null,
  comments: [],

  fetchBoard: async (boardId) => {
    set({ isLoading: true });
    try {
      const data = await apiFetch<BoardDetail>(`/boards/${boardId}`, {
        tenantId: getTenantId(),
      });
      set({
        boardId: data.id,
        boardName: data.name,
        columns: data.columns.map((c) => ({
          ...c,
          tasks: [...c.tasks].sort((a, b) => a.position - b.position),
        })),
        isLoading: false,
      });
    } catch {
      set({ isLoading: false });
    }
  },

  setActiveTask: (task) => set({ activeTask: task }),

  moveTask: async (taskId, fromColumnId, toColumnId, newPosition) => {
    const prev = get().columns;

    // Optimistic update
    set((state) => {
      const cols = structuredClone(state.columns);
      const fromCol = cols.find((c) => c.id === fromColumnId);
      const toCol = cols.find((c) => c.id === toColumnId);
      if (!fromCol || !toCol) return state;

      const taskIdx = fromCol.tasks.findIndex((t) => t.id === taskId);
      if (taskIdx === -1) return state;
      const [task] = fromCol.tasks.splice(taskIdx, 1);

      task.position = newPosition;
      toCol.tasks.splice(newPosition, 0, task);
      toCol.tasks.forEach((t, i) => (t.position = i));
      if (fromColumnId !== toColumnId) {
        fromCol.tasks.forEach((t, i) => (t.position = i));
      }

      return { columns: cols };
    });

    try {
      await apiFetch("/tasks/move", {
        method: "PATCH",
        body: JSON.stringify({
          task_id: taskId,
          to_column_id: toColumnId,
          new_position: newPosition,
        }),
        tenantId: getTenantId(),
      });
    } catch {
      set({ columns: prev });
      throw new Error("move_failed");
    }
  },

  createTask: async (columnId, title) => {
    const tempId = `temp-${Date.now()}`;
    const col = get().columns.find((c) => c.id === columnId);
    const position = col ? col.tasks.length : 0;

    // Optimistic
    set((state) => {
      const cols = state.columns.map((c) => {
        if (c.id !== columnId) return c;
        return {
          ...c,
          tasks: [
            ...c.tasks,
            { id: tempId, title, position, priority: "medium", due_date: null, assignee: null, labels: [] },
          ],
        };
      });
      return { columns: cols };
    });

    try {
      const task = await apiFetch<TaskDetail>("/tasks", {
        method: "POST",
        body: JSON.stringify({ column_id: columnId, title }),
        tenantId: getTenantId(),
      });
      // Replace temp with real
      set((state) => {
        const cols = state.columns.map((c) => {
          if (c.id !== columnId) return c;
          return {
            ...c,
            tasks: c.tasks.map((t) =>
              t.id === tempId
                ? { id: task.id, title: task.title, position: task.position, priority: task.priority, due_date: task.due_date, assignee: task.assignee, labels: task.labels }
                : t
            ),
          };
        });
        return { columns: cols };
      });
    } catch {
      // Rollback
      set((state) => ({
        columns: state.columns.map((c) => {
          if (c.id !== columnId) return c;
          return { ...c, tasks: c.tasks.filter((t) => t.id !== tempId) };
        }),
      }));
      throw new Error("create_failed");
    }
  },

  updateTask: async (taskId, data) => {
    try {
      const updated = await apiFetch<TaskDetail>(`/tasks/${taskId}`, {
        method: "PATCH",
        body: JSON.stringify(data),
        tenantId: getTenantId(),
      });
      set((state) => ({
        columns: state.columns.map((c) => ({
          ...c,
          tasks: c.tasks.map((t) =>
            t.id === taskId
              ? { ...t, title: updated.title, priority: updated.priority, due_date: updated.due_date, assignee: updated.assignee, labels: updated.labels }
              : t
          ),
        })),
        selectedTask: state.selectedTask?.id === taskId ? updated : state.selectedTask,
      }));
    } catch {
      throw new Error("update_failed");
    }
  },

  deleteTask: async (taskId, columnId) => {
    set((state) => ({
      columns: state.columns.map((c) => {
        if (c.id !== columnId) return c;
        return { ...c, tasks: c.tasks.filter((t) => t.id !== taskId) };
      }),
      selectedTask: state.selectedTask?.id === taskId ? null : state.selectedTask,
    }));
    try {
      await apiFetch(`/tasks/${taskId}`, {
        method: "DELETE",
        tenantId: getTenantId(),
      });
    } catch {
      // Refetch on failure
      const boardId = get().boardId;
      if (boardId) get().fetchBoard(boardId);
      throw new Error("delete_failed");
    }
  },

  createColumn: async (boardId, name) => {
    try {
      const col = await apiFetch<Column & { board_id: string }>("/columns", {
        method: "POST",
        body: JSON.stringify({ board_id: boardId, name }),
        tenantId: getTenantId(),
      });
      set((state) => ({
        columns: [...state.columns, { id: col.id, name: col.name, position: col.position, color: col.color, wip_limit: col.wip_limit, tasks: [] }],
      }));
    } catch {
      throw new Error("create_failed");
    }
  },

  updateColumn: async (columnId, data) => {
    try {
      const updated = await apiFetch<Column & { board_id: string }>(`/columns/${columnId}`, {
        method: "PATCH",
        body: JSON.stringify(data),
        tenantId: getTenantId(),
      });
      set((state) => ({
        columns: state.columns.map((c) =>
          c.id === columnId ? { ...c, name: updated.name, color: updated.color, wip_limit: updated.wip_limit } : c
        ),
      }));
    } catch {
      throw new Error("update_failed");
    }
  },

  deleteColumn: async (columnId) => {
    const prev = get().columns;
    set((state) => ({
      columns: state.columns.filter((c) => c.id !== columnId),
    }));
    try {
      await apiFetch(`/columns/${columnId}`, {
        method: "DELETE",
        tenantId: getTenantId(),
      });
    } catch {
      set({ columns: prev });
      throw new Error("delete_failed");
    }
  },

  fetchTaskDetail: async (taskId) => {
    try {
      const task = await apiFetch<TaskDetail>(`/tasks/${taskId}`, {
        tenantId: getTenantId(),
      });
      set({ selectedTask: task });
    } catch {
      // ignore
    }
  },

  fetchComments: async (taskId) => {
    try {
      const comments = await apiFetch<Comment[]>(`/tasks/${taskId}/comments`, {
        tenantId: getTenantId(),
      });
      set({ comments: comments || [] });
    } catch {
      set({ comments: [] });
    }
  },

  addComment: async (taskId, content) => {
    const comment = await apiFetch<Comment>(`/tasks/${taskId}/comments`, {
      method: "POST",
      body: JSON.stringify({ content }),
      tenantId: getTenantId(),
    });
    set((state) => ({ comments: [...state.comments, comment] }));
  },

  clearSelectedTask: () => set({ selectedTask: null, comments: [] }),

  handleWSMessage: (message) => {
    const { type, payload } = message;
    const p = payload as Record<string, unknown>;

    switch (type) {
      case "task:created": {
        const task = p as unknown as TaskDetail;
        set((state) => {
          const exists = state.columns.some((c) => c.tasks.some((t) => t.id === task.id));
          if (exists) return state;
          return {
            columns: state.columns.map((c) => {
              if (c.id !== task.column_id) return c;
              return {
                ...c,
                tasks: [...c.tasks, {
                  id: task.id, title: task.title, position: task.position,
                  priority: task.priority, due_date: task.due_date,
                  assignee: task.assignee, labels: task.labels,
                }],
              };
            }),
          };
        });
        break;
      }
      case "task:updated": {
        const task = p as unknown as TaskDetail;
        set((state) => ({
          columns: state.columns.map((c) => ({
            ...c,
            tasks: c.tasks.map((t) =>
              t.id === task.id
                ? { ...t, title: task.title, priority: task.priority, due_date: task.due_date, assignee: task.assignee, labels: task.labels }
                : t
            ),
          })),
          selectedTask: state.selectedTask?.id === task.id ? { ...state.selectedTask, ...task } : state.selectedTask,
        }));
        break;
      }
      case "task:deleted": {
        const { id } = p as { id: string };
        set((state) => ({
          columns: state.columns.map((c) => ({
            ...c,
            tasks: c.tasks.filter((t) => t.id !== id),
          })),
          selectedTask: state.selectedTask?.id === id ? null : state.selectedTask,
        }));
        break;
      }
      case "task:moved": {
        // Refetch to get correct state
        const boardId = get().boardId;
        if (boardId) get().fetchBoard(boardId);
        break;
      }
      case "column:created":
      case "column:updated":
      case "column:deleted":
      case "column:reordered": {
        const boardId = get().boardId;
        if (boardId) get().fetchBoard(boardId);
        break;
      }
    }
  },
}));
