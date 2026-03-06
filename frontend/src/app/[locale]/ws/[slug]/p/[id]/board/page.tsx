"use client";

import { useEffect } from "react";
import { useParams } from "next/navigation";
import { useTranslations } from "next-intl";
import { useBoardStore } from "@/stores/boardStore";
import { useWebSocket } from "@/hooks/useWebSocket";
import { KanbanBoard } from "@/components/board/KanbanBoard";

export default function BoardPage() {
  const t = useTranslations("common");
  const params = useParams();
  const boardId = params?.id as string;
  const { fetchBoard, isLoading, boardName } = useBoardStore();

  useEffect(() => {
    if (boardId) fetchBoard(boardId);
  }, [boardId, fetchBoard]);

  useWebSocket(boardId);

  if (isLoading) {
    return (
      <div className="flex items-center justify-center py-20">
        <p className="text-muted-foreground">{t("loading")}</p>
      </div>
    );
  }

  return (
    <div className="flex flex-col gap-4">
      {boardName && (
        <h1 className="text-xl font-bold">{boardName}</h1>
      )}
      <KanbanBoard />
    </div>
  );
}
