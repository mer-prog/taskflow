"use client";

import { create } from "zustand";
import { apiFetch, setAccessToken } from "@/lib/api";

interface User {
  id: string;
  email: string;
  display_name: string;
  avatar_url?: string;
  created_at: string;
}

interface AuthResponse {
  access_token: string;
  user: User;
}

interface AuthState {
  user: User | null;
  accessToken: string | null;
  tenantId: string | null;
  isAuthenticated: boolean;
  login: (email: string, password: string) => Promise<void>;
  register: (email: string, password: string, displayName: string) => Promise<void>;
  logout: () => Promise<void>;
  setTenantId: (tenantId: string) => void;
  setAuth: (user: User, token: string) => void;
}

export const useAuthStore = create<AuthState>((set) => ({
  user: null,
  accessToken: null,
  tenantId: null,
  isAuthenticated: false,

  login: async (email, password) => {
    const data = await apiFetch<AuthResponse>("/auth/login", {
      method: "POST",
      body: JSON.stringify({ email, password }),
    });
    setAccessToken(data.access_token);
    set({
      user: data.user,
      accessToken: data.access_token,
      isAuthenticated: true,
    });
  },

  register: async (email, password, displayName) => {
    const data = await apiFetch<AuthResponse>("/auth/register", {
      method: "POST",
      body: JSON.stringify({ email, password, display_name: displayName }),
    });
    setAccessToken(data.access_token);
    set({
      user: data.user,
      accessToken: data.access_token,
      isAuthenticated: true,
    });
  },

  logout: async () => {
    try {
      await apiFetch("/auth/logout", { method: "POST" });
    } catch {
      // ignore logout errors
    }
    setAccessToken(null);
    set({
      user: null,
      accessToken: null,
      tenantId: null,
      isAuthenticated: false,
    });
  },

  setTenantId: (tenantId) => set({ tenantId }),

  setAuth: (user, token) => {
    setAccessToken(token);
    set({ user, accessToken: token, isAuthenticated: true });
  },
}));
