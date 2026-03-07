"use client";

import { useCallback, useMemo, useState } from "react";
import { useTranslations } from "next-intl";
import {
  DndContext,
  DragOverlay,
  closestCorners,
  PointerSensor,
  TouchSensor,
  useSensor,
  useSensors,
  type DragStartEvent,
  type DragEndEvent,
  type DragOverEvent,
} from "@dnd-kit/core";
import { SortableContext, horizontalListSortingStrategy } from "@dnd-kit/sortable";
import { useBoardStore } from "@/stores/boardStore";
import { KanbanColumn } from "./KanbanColumn";
import { TaskCard } from "./TaskCard";
import { AddColumnButton } from "./AddColumnButton";
import { TaskDetailDrawer } from "./TaskDetailDrawer";
import { toast } from "sonner";

export function KanbanBoard() {
  const t = useTranslations("board");
  const { columns, activeTask, setActiveTask, moveTask, selectedTask } = useBoardStore();
  const [overId, setOverId] = useState<string | null>(null);

  const columnIds = useMemo(() => columns.map((c) => c.id), [columns]);

  const sensors = useSensors(
    useSensor(PointerSensor, { activationConstraint: { distance: 5 } }),
    useSensor(TouchSensor, { activationConstraint: { delay: 200, tolerance: 5 } })
  );

  const findColumnByTaskId = useCallback(
    (taskId: string) => columns.find((c) => c.tasks.some((t) => t.id === taskId)),
    [columns]
  );

  function handleDragStart(event: DragStartEvent) {
    const { active } = event;
    const col = findColumnByTaskId(active.id as string);
    const task = col?.tasks.find((t) => t.id === active.id);
    if (task) setActiveTask(task);
  }

  function handleDragOver(event: DragOverEvent) {
    setOverId((event.over?.id as string) || null);
  }

  async function handleDragEnd(event: DragEndEvent) {
    const { active, over } = event;
    setActiveTask(null);
    setOverId(null);

    if (!over) return;

    const activeId = active.id as string;
    const overId = over.id as string;

    const fromCol = findColumnByTaskId(activeId);
    if (!fromCol) return;

    // Determine target column and position
    let toCol = findColumnByTaskId(overId);
    let newPosition: number;

    if (toCol) {
      // Dropped on a task
      const overTask = toCol.tasks.find((t) => t.id === overId);
      newPosition = overTask ? overTask.position : toCol.tasks.length;
    } else {
      // Dropped on a column directly
      toCol = columns.find((c) => c.id === overId);
      if (!toCol) return;
      newPosition = toCol.tasks.length;
    }

    if (fromCol.id === toCol.id && fromCol.tasks.find((t) => t.id === activeId)?.position === newPosition) {
      return;
    }

    try {
      await moveTask(activeId, fromCol.id, toCol.id, newPosition);
    } catch {
      toast.error(t("moveFailed"));
    }
  }

  return (
    <>
      <DndContext
        sensors={sensors}
        collisionDetection={closestCorners}
        onDragStart={handleDragStart}
        onDragOver={handleDragOver}
        onDragEnd={handleDragEnd}
      >
        <div className="flex gap-4 overflow-x-auto pb-4 kanban-scroll">
          <SortableContext
            items={columnIds}
            strategy={horizontalListSortingStrategy}
          >
            {columns.map((column) => (
              <KanbanColumn key={column.id} column={column} isOver={overId === column.id} />
            ))}
          </SortableContext>
          <AddColumnButton />
        </div>

        <DragOverlay dropAnimation={null}>
          {activeTask ? (
            <div className="rotate-[3deg] opacity-90">
              <TaskCard task={activeTask} isOverlay />
            </div>
          ) : null}
        </DragOverlay>
      </DndContext>

      {selectedTask && <TaskDetailDrawer />}
    </>
  );
}
