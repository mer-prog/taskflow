"use client";

import { create } from "zustand";
import { apiFetch } from "@/lib/api";
import { useAuthStore } from "./authStore";

export interface Tenant {
  id: string;
  name: string;
  slug: string;
  created_at: string;
  updated_at: string;
}

export interface Project {
  id: string;
  tenant_id: string;
  name: string;
  description?: string;
  created_at: string;
  updated_at: string;
}

export interface Member {
  user_id: string;
  email: string;
  display_name: string;
  avatar_url?: string;
  role: string;
  joined_at: string;
}

interface WorkspaceState {
  tenant: Tenant | null;
  projects: Project[];
  members: Member[];
  isLoading: boolean;
  currentUserRole: string | null;

  fetchTenants: () => Promise<Tenant[]>;
  fetchTenant: (tenantId: string) => Promise<void>;
  updateTenant: (tenantId: string, name: string) => Promise<void>;

  fetchProjects: () => Promise<void>;
  createProject: (name: string, description: string) => Promise<Project>;
  deleteProject: (projectId: string) => Promise<void>;

  fetchMembers: (tenantId: string) => Promise<void>;
  updateMemberRole: (tenantId: string, userId: string, role: string) => Promise<void>;
  removeMember: (tenantId: string, userId: string) => Promise<void>;
}

function getTenantId() {
  return useAuthStore.getState().tenantId || "";
}

function getCurrentUserId() {
  return useAuthStore.getState().user?.id || "";
}

export const useWorkspaceStore = create<WorkspaceState>((set, get) => ({
  tenant: null,
  projects: [],
  members: [],
  isLoading: false,
  currentUserRole: null,

  fetchTenants: async () => {
    const data = await apiFetch<Tenant[]>("/tenants");
    return data || [];
  },

  fetchTenant: async (tenantId) => {
    set({ isLoading: true });
    try {
      const data = await apiFetch<Tenant>(`/tenants/${tenantId}`);
      set({ tenant: data, isLoading: false });
    } catch {
      set({ isLoading: false });
    }
  },

  updateTenant: async (tenantId, name) => {
    const data = await apiFetch<Tenant>(`/tenants/${tenantId}`, {
      method: "PATCH",
      body: JSON.stringify({ name }),
    });
    set({ tenant: data });
  },

  fetchProjects: async () => {
    try {
      const data = await apiFetch<Project[]>("/projects", {
        tenantId: getTenantId(),
      });
      set({ projects: data || [] });
    } catch {
      set({ projects: [] });
    }
  },

  createProject: async (name, description) => {
    const data = await apiFetch<Project>("/projects", {
      method: "POST",
      body: JSON.stringify({ name, description }),
      tenantId: getTenantId(),
    });
    set((state) => ({ projects: [...state.projects, data] }));
    return data;
  },

  deleteProject: async (projectId) => {
    await apiFetch(`/projects/${projectId}`, {
      method: "DELETE",
      tenantId: getTenantId(),
    });
    set((state) => ({
      projects: state.projects.filter((p) => p.id !== projectId),
    }));
  },

  fetchMembers: async (tenantId) => {
    try {
      const data = await apiFetch<Member[]>(`/tenants/${tenantId}/members`);
      const currentUserId = getCurrentUserId();
      const currentMember = data?.find((m) => m.user_id === currentUserId);
      set({
        members: data || [],
        currentUserRole: currentMember?.role || null,
      });
    } catch {
      set({ members: [], currentUserRole: null });
    }
  },

  updateMemberRole: async (tenantId, userId, role) => {
    await apiFetch(`/tenants/${tenantId}/members/${userId}`, {
      method: "PATCH",
      body: JSON.stringify({ role }),
    });
    set((state) => ({
      members: state.members.map((m) =>
        m.user_id === userId ? { ...m, role } : m
      ),
    }));
  },

  removeMember: async (tenantId, userId) => {
    await apiFetch(`/tenants/${tenantId}/members/${userId}`, {
      method: "DELETE",
    });
    set((state) => ({
      members: state.members.filter((m) => m.user_id !== userId),
    }));
  },
}));
