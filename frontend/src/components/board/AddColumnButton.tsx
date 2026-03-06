"use client";

import { useState, useRef } from "react";
import { useTranslations } from "next-intl";
import { useBoardStore } from "@/stores/boardStore";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { toast } from "sonner";

export function AddColumnButton() {
  const t = useTranslations("board");
  const { boardId, createColumn } = useBoardStore();
  const [isOpen, setIsOpen] = useState(false);
  const [name, setName] = useState("");
  const inputRef = useRef<HTMLInputElement>(null);

  async function handleSubmit() {
    const trimmed = name.trim();
    if (!trimmed || !boardId) return;
    setName("");
    setIsOpen(false);
    try {
      await createColumn(boardId, trimmed);
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
      setName("");
    }
  }

  if (!isOpen) {
    return (
      <Button
        variant="outline"
        className="h-auto w-[85vw] shrink-0 justify-start rounded-lg border-dashed py-6 text-muted-foreground sm:w-[300px] min-h-[44px]"
        onClick={() => {
          setIsOpen(true);
          setTimeout(() => inputRef.current?.focus(), 0);
        }}
      >
        <svg width="16" height="16" viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="2" className="mr-2">
          <path d="M8 3v10M3 8h10" />
        </svg>
        {t("addColumn")}
      </Button>
    );
  }

  return (
    <div className="w-[85vw] shrink-0 sm:w-[300px]">
      <Input
        ref={inputRef}
        value={name}
        onChange={(e) => setName(e.target.value)}
        onKeyDown={handleKeyDown}
        onBlur={() => {
          if (!name.trim()) setIsOpen(false);
        }}
        placeholder={t("columnName")}
        className="min-h-[44px]"
        autoFocus
      />
    </div>
  );
}
