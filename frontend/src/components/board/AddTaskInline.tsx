"use client";

import { useState, useRef } from "react";
import { useTranslations } from "next-intl";
import { useBoardStore } from "@/stores/boardStore";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { toast } from "sonner";

export function AddTaskInline({ columnId }: { columnId: string }) {
  const t = useTranslations("board");
  const createTask = useBoardStore((s) => s.createTask);
  const [isOpen, setIsOpen] = useState(false);
  const [title, setTitle] = useState("");
  const inputRef = useRef<HTMLInputElement>(null);

  async function handleSubmit() {
    const trimmed = title.trim();
    if (!trimmed) return;
    setTitle("");
    try {
      await createTask(columnId, trimmed);
      // Keep focus for continuous add
      inputRef.current?.focus();
    } catch {
      toast.error(t("createFailed"));
    }
  }

  function handleKeyDown(e: React.KeyboardEvent) {
    if (e.key === "Enter") {
      e.preventDefault();
      handleSubmit();
    }
    if (e.key === "Escape") {
      setIsOpen(false);
      setTitle("");
    }
  }

  if (!isOpen) {
    return (
      <Button
        variant="ghost"
        size="sm"
        className="w-full justify-start text-muted-foreground min-h-[44px]"
        onClick={() => {
          setIsOpen(true);
          setTimeout(() => inputRef.current?.focus(), 0);
        }}
      >
        <svg width="16" height="16" viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="2" className="mr-1">
          <path d="M8 3v10M3 8h10" />
        </svg>
        {t("addTask")}
      </Button>
    );
  }

  return (
    <div className="space-y-2">
      <Input
        ref={inputRef}
        value={title}
        onChange={(e) => setTitle(e.target.value)}
        onKeyDown={handleKeyDown}
        onBlur={() => {
          if (!title.trim()) setIsOpen(false);
        }}
        placeholder={t("taskTitle")}
        className="min-h-[44px]"
        autoFocus
      />
    </div>
  );
}
