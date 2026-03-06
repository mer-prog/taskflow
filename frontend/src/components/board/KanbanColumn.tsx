"use client";

import { useTranslations } from "next-intl";
import { useDroppable } from "@dnd-kit/core";
import { SortableContext, verticalListSortingStrategy } from "@dnd-kit/sortable";
import { type Column } from "@/stores/boardStore";
import { TaskCard } from "./TaskCard";
import { AddTaskInline } from "./AddTaskInline";
import { Badge } from "@/components/ui/badge";
import { cn } from "@/lib/utils";

interface KanbanColumnProps {
  column: Column;
  isOver?: boolean;
}

export function KanbanColumn({ column, isOver }: KanbanColumnProps) {
  const t = useTranslations("board");
  const { setNodeRef } = useDroppable({ id: column.id });
  const isWipExceeded = column.wip_limit != null && column.tasks.length > column.wip_limit;

  return (
    <div
      ref={setNodeRef}
      className={cn(
        "flex w-[85vw] shrink-0 flex-col rounded-lg bg-muted/50 sm:w-[300px]",
        isOver && "ring-2 ring-primary/30"
      )}
    >
      {/* Column header */}
      <div className="flex items-center gap-2 px-3 py-3">
        <div
          className="h-3 w-3 shrink-0 rounded-full"
          style={{ backgroundColor: column.color }}
        />
        <h3 className="flex-1 truncate text-sm font-semibold">{column.name}</h3>
        <Badge variant="secondary" className="text-xs tabular-nums">
          {column.tasks.length}
          {column.wip_limit != null && `/${column.wip_limit}`}
        </Badge>
        {isWipExceeded && (
          <Badge variant="destructive" className="text-xs">
            {t("wipLimitExceeded")}
          </Badge>
        )}
      </div>

      {/* Tasks */}
      <div className="flex min-h-[60px] flex-1 flex-col gap-2 px-2 pb-2">
        <SortableContext
          items={column.tasks.map((t) => t.id)}
          strategy={verticalListSortingStrategy}
        >
          {column.tasks.map((task) => (
            <TaskCard key={task.id} task={task} />
          ))}
        </SortableContext>
      </div>

      {/* Add task */}
      <div className="px-2 pb-2">
        <AddTaskInline columnId={column.id} />
      </div>
    </div>
  );
}
