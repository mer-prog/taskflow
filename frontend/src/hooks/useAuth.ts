"use client";

import { useEffect } from "react";
import { useAuthStore } from "@/stores/authStore";
import { useRouter } from "@/i18n/navigation";

export function useRequireAuth() {
  const { isAuthenticated } = useAuthStore();
  const router = useRouter();

  useEffect(() => {
    if (!isAuthenticated) {
      router.replace("/login");
    }
  }, [isAuthenticated, router]);

  return { isAuthenticated };
}

export function useRedirectIfAuth() {
  const { isAuthenticated } = useAuthStore();
  const router = useRouter();

  useEffect(() => {
    if (isAuthenticated) {
      router.replace("/");
    }
  }, [isAuthenticated, router]);
}
