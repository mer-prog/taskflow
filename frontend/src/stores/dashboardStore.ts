"use client";

import { create } from "zustand";
import { apiFetch } from "@/lib/api";
import { useAuthStore } from "./authStore";

interface ColumnCount {
  column_id: string;
  column_name: string;
  board_name: string;
  task_count: number;
}

export interface DashboardSummary {
  total_tasks: number;
  overdue_tasks: number;
  by_column: ColumnCount[];
}

export interface DashboardTask {
  id: string;
  title: string;
  priority: string;
  due_date: string | null;
  column_name: string;
  board_name?: string;
  assignee_name?: string;
}

interface DashboardState {
  summary: DashboardSummary | null;
  myTasks: DashboardTask[];
  overdueTasks: DashboardTask[];
  isLoading: boolean;
  fetchSummary: () => Promise<void>;
  fetchMyTasks: () => Promise<void>;
  fetchOverdueTasks: () => Promise<void>;
  fetchAll: () => Promise<void>;
}

function getTenantId() {
  return useAuthStore.getState().tenantId || "";
}

export const useDashboardStore = create<DashboardState>((set, get) => ({
  summary: null,
  myTasks: [],
  overdueTasks: [],
  isLoading: false,

  fetchSummary: async () => {
    try {
      const data = await apiFetch<DashboardSummary>("/dashboard/summary", {
        tenantId: getTenantId(),
      });
      set({ summary: data });
    } catch {
      // ignore
    }
  },

  fetchMyTasks: async () => {
    try {
      const data = await apiFetch<DashboardTask[]>("/dashboard/my-tasks", {
        tenantId: getTenantId(),
      });
      set({ myTasks: data || [] });
    } catch {
      set({ myTasks: [] });
    }
  },

  fetchOverdueTasks: async () => {
    try {
      const data = await apiFetch<DashboardTask[]>("/dashboard/overdue", {
        tenantId: getTenantId(),
      });
      set({ overdueTasks: data || [] });
    } catch {
      set({ overdueTasks: [] });
    }
  },

  fetchAll: async () => {
    set({ isLoading: true });
    await Promise.all([
      get().fetchSummary(),
      get().fetchMyTasks(),
      get().fetchOverdueTasks(),
    ]);
    set({ isLoading: false });
  },
}));
