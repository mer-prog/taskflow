"use client";

import { useTranslations } from "next-intl";
import { useSortable } from "@dnd-kit/sortable";
import { CSS } from "@dnd-kit/utilities";
import { type Task, useBoardStore } from "@/stores/boardStore";
import { Badge } from "@/components/ui/badge";
import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import { cn } from "@/lib/utils";

const priorityConfig: Record<string, { color: string; variant: "default" | "secondary" | "destructive" | "outline" }> = {
  low: { color: "bg-gray-100 text-gray-600", variant: "secondary" },
  medium: { color: "bg-blue-100 text-blue-700", variant: "secondary" },
  high: { color: "bg-orange-100 text-orange-700", variant: "secondary" },
  critical: { color: "bg-red-100 text-red-700", variant: "destructive" },
};

interface TaskCardProps {
  task: Task;
  isOverlay?: boolean;
}

export function TaskCard({ task, isOverlay }: TaskCardProps) {
  const t = useTranslations("board");
  const fetchTaskDetail = useBoardStore((s) => s.fetchTaskDetail);
  const fetchComments = useBoardStore((s) => s.fetchComments);

  const {
    attributes,
    listeners,
    setNodeRef,
    transform,
    transition,
    isDragging,
  } = useSortable({ id: task.id });

  const style = {
    transform: CSS.Transform.toString(transform),
    transition,
  };

  const dueInfo = getDueInfo(task.due_date);
  const pCfg = priorityConfig[task.priority] || priorityConfig.medium;

  function handleClick() {
    if (isDragging) return;
    fetchTaskDetail(task.id);
    fetchComments(task.id);
  }

  const initials = task.assignee?.display_name
    ?.split(" ")
    .map((n) => n[0])
    .join("")
    .toUpperCase()
    .slice(0, 2) || "";

  return (
    <div
      ref={setNodeRef}
      style={style}
      className={cn(
        "group cursor-pointer rounded-lg border bg-card p-3 shadow-sm transition-all hover:shadow-md hover:-translate-y-0.5",
        isDragging && "opacity-30",
        isOverlay && "shadow-lg"
      )}
      onClick={handleClick}
    >
      {/* Drag handle for mobile */}
      <div className="flex items-start gap-2">
        <button
          {...attributes}
          {...listeners}
          className="mt-0.5 shrink-0 touch-none text-muted-foreground opacity-0 transition-opacity group-hover:opacity-100 md:opacity-0"
          aria-label="Drag"
        >
          <svg width="14" height="14" viewBox="0 0 14 14" fill="currentColor">
            <circle cx="4" cy="3" r="1.5" />
            <circle cx="10" cy="3" r="1.5" />
            <circle cx="4" cy="7" r="1.5" />
            <circle cx="10" cy="7" r="1.5" />
            <circle cx="4" cy="11" r="1.5" />
            <circle cx="10" cy="11" r="1.5" />
          </svg>
        </button>
        <div className="min-w-0 flex-1">
          <p className="text-sm font-medium leading-snug">{task.title}</p>
        </div>
      </div>

      {/* Labels */}
      {task.labels.length > 0 && (
        <div className="mt-2 flex flex-wrap gap-1">
          {task.labels.map((label) => (
            <span
              key={label.id}
              className="inline-block rounded-full px-2 py-0.5 text-[10px] font-medium text-white"
              style={{ backgroundColor: label.color }}
            >
              {label.name}
            </span>
          ))}
        </div>
      )}

      {/* Footer: priority, due date, assignee */}
      <div className="mt-2 flex items-center gap-1.5">
        <Badge className={cn("text-[10px] px-1.5 py-0", pCfg.color)} variant={pCfg.variant}>
          {t(`priority${task.priority.charAt(0).toUpperCase() + task.priority.slice(1)}` as "priorityLow")}
        </Badge>

        {dueInfo && (
          <Badge
            variant="outline"
            className={cn(
              "text-[10px] px-1.5 py-0",
              dueInfo.status === "overdue" && "border-red-300 bg-red-50 text-red-700",
              dueInfo.status === "soon" && "border-yellow-300 bg-yellow-50 text-yellow-700"
            )}
          >
            {dueInfo.label}
          </Badge>
        )}

        {task.assignee && (
          <Avatar className="ml-auto h-5 w-5">
            <AvatarFallback className="text-[9px]">{initials}</AvatarFallback>
          </Avatar>
        )}
      </div>
    </div>
  );
}

function getDueInfo(dueDate: string | null): { status: "overdue" | "soon" | "normal"; label: string } | null {
  if (!dueDate) return null;
  const due = new Date(dueDate);
  const now = new Date();
  const diffMs = due.getTime() - now.getTime();
  const diffDays = Math.ceil(diffMs / (1000 * 60 * 60 * 24));

  const formatted = due.toLocaleDateString(undefined, { month: "short", day: "numeric" });

  if (diffDays < 0) return { status: "overdue", label: formatted };
  if (diffDays <= 3) return { status: "soon", label: formatted };
  return { status: "normal", label: formatted };
}
