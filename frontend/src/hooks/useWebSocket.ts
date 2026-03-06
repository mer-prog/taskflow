"use client";

import { useEffect, useRef, useCallback } from "react";
import { useBoardStore } from "@/stores/boardStore";
import { getAccessToken } from "@/lib/api";
import { useAuthStore } from "@/stores/authStore";

const WS_URL = process.env.NEXT_PUBLIC_WS_URL || "ws://localhost:8080/api/v1/ws";
const MAX_BACKOFF = 30000;

export function useWebSocket(boardId: string | null) {
  const wsRef = useRef<WebSocket | null>(null);
  const backoffRef = useRef(1000);
  const mountedRef = useRef(true);
  const handleWSMessage = useBoardStore((s) => s.handleWSMessage);
  const userId = useAuthStore((s) => s.user?.id);

  const connect = useCallback(() => {
    if (!boardId || !mountedRef.current) return;

    const token = getAccessToken();
    if (!token) return;

    const ws = new WebSocket(`${WS_URL}?board_id=${boardId}&token=${token}`);
    wsRef.current = ws;

    ws.onopen = () => {
      backoffRef.current = 1000;
    };

    ws.onmessage = (event) => {
      try {
        const msg = JSON.parse(event.data);
        // Skip own events
        if (msg.user_id === userId) return;
        const payload = typeof msg.payload === "string" ? JSON.parse(msg.payload) : msg.payload;
        handleWSMessage({ type: msg.type, payload, user_id: msg.user_id });
      } catch {
        // ignore parse errors
      }
    };

    ws.onclose = () => {
      if (!mountedRef.current) return;
      const delay = backoffRef.current;
      backoffRef.current = Math.min(delay * 2, MAX_BACKOFF);
      setTimeout(() => {
        if (mountedRef.current) connect();
      }, delay);
    };

    ws.onerror = () => {
      ws.close();
    };
  }, [boardId, handleWSMessage, userId]);

  useEffect(() => {
    mountedRef.current = true;
    connect();

    return () => {
      mountedRef.current = false;
      wsRef.current?.close();
      wsRef.current = null;
    };
  }, [connect]);
}
