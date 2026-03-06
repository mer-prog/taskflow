"use client";

import { useState } from "react";
import { useTranslations } from "next-intl";
import { useBoardStore } from "@/stores/boardStore";
import { Sheet, SheetContent, SheetHeader, SheetTitle } from "@/components/ui/sheet";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { Label } from "@/components/ui/label";
import { Badge } from "@/components/ui/badge";
import { Separator } from "@/components/ui/separator";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import { toast } from "sonner";

const priorities = ["low", "medium", "high", "critical"] as const;

export function TaskDetailDrawer() {
  const t = useTranslations("board");
  const tc = useTranslations("common");
  const { selectedTask, comments, clearSelectedTask, updateTask, deleteTask, addComment } = useBoardStore();
  const [commentText, setCommentText] = useState("");
  const [editingTitle, setEditingTitle] = useState(false);
  const [titleValue, setTitleValue] = useState("");
  const [descValue, setDescValue] = useState("");
  const [isSavingDesc, setIsSavingDesc] = useState(false);

  if (!selectedTask) return null;

  async function handleTitleSave() {
    if (!selectedTask || !titleValue.trim()) return;
    setEditingTitle(false);
    try {
      await updateTask(selectedTask.id, { title: titleValue.trim() });
    } catch {
      toast.error(t("updateFailed"));
    }
  }

  async function handleDescSave() {
    if (!selectedTask) return;
    setIsSavingDesc(true);
    try {
      await updateTask(selectedTask.id, { description: descValue });
    } catch {
      toast.error(t("updateFailed"));
    } finally {
      setIsSavingDesc(false);
    }
  }

  async function handlePriorityChange(value: string) {
    if (!selectedTask) return;
    try {
      await updateTask(selectedTask.id, { priority: value });
    } catch {
      toast.error(t("updateFailed"));
    }
  }

  async function handleDelete() {
    if (!selectedTask) return;
    clearSelectedTask();
    try {
      await deleteTask(selectedTask.id, selectedTask.column_id);
    } catch {
      toast.error(t("deleteFailed"));
    }
  }

  async function handleAddComment() {
    if (!selectedTask || !commentText.trim()) return;
    const text = commentText.trim();
    setCommentText("");
    try {
      await addComment(selectedTask.id, text);
    } catch {
      toast.error(t("createFailed"));
    }
  }

  return (
    <Sheet open={!!selectedTask} onOpenChange={(open) => { if (!open) clearSelectedTask(); }}>
      <SheetContent className="w-full overflow-y-auto sm:max-w-lg">
        <SheetHeader>
          <SheetTitle className="sr-only">{t("task")}</SheetTitle>
        </SheetHeader>

        <div className="space-y-6 pt-2">
          {/* Title */}
          {editingTitle ? (
            <Input
              value={titleValue}
              onChange={(e) => setTitleValue(e.target.value)}
              onBlur={handleTitleSave}
              onKeyDown={(e) => { if (e.key === "Enter") handleTitleSave(); if (e.key === "Escape") setEditingTitle(false); }}
              className="text-lg font-semibold min-h-[44px]"
              autoFocus
            />
          ) : (
            <h2
              className="cursor-pointer text-lg font-semibold hover:bg-muted/50 rounded px-1 py-0.5 -mx-1"
              onClick={() => { setTitleValue(selectedTask.title); setEditingTitle(true); }}
            >
              {selectedTask.title}
            </h2>
          )}

          {/* Priority */}
          <div className="space-y-2">
            <Label>{t("priority")}</Label>
            <Select value={selectedTask.priority} onValueChange={handlePriorityChange}>
              <SelectTrigger className="min-h-[44px]">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                {priorities.map((p) => (
                  <SelectItem key={p} value={p}>
                    {t(`priority${p.charAt(0).toUpperCase() + p.slice(1)}` as "priorityLow")}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>

          {/* Assignee */}
          {selectedTask.assignee && (
            <div className="space-y-2">
              <Label>{t("assignee")}</Label>
              <div className="flex items-center gap-2">
                <Avatar className="h-6 w-6">
                  <AvatarFallback className="text-[10px]">
                    {selectedTask.assignee.display_name?.charAt(0).toUpperCase() || "?"}
                  </AvatarFallback>
                </Avatar>
                <span className="text-sm">{selectedTask.assignee.display_name}</span>
              </div>
            </div>
          )}

          {/* Due date */}
          {selectedTask.due_date && (
            <div className="space-y-2">
              <Label>{t("dueDate")}</Label>
              <p className="text-sm">{new Date(selectedTask.due_date).toLocaleDateString()}</p>
            </div>
          )}

          {/* Labels */}
          {selectedTask.labels.length > 0 && (
            <div className="space-y-2">
              <Label>{t("labels")}</Label>
              <div className="flex flex-wrap gap-1">
                {selectedTask.labels.map((label) => (
                  <Badge
                    key={label.id}
                    className="text-white text-xs"
                    style={{ backgroundColor: label.color }}
                  >
                    {label.name}
                  </Badge>
                ))}
              </div>
            </div>
          )}

          {/* Description */}
          <div className="space-y-2">
            <Label>{t("description")}</Label>
            <Textarea
              value={descValue || selectedTask.description || ""}
              onChange={(e) => setDescValue(e.target.value)}
              onBlur={handleDescSave}
              placeholder={t("description")}
              className="min-h-[100px] resize-none"
              disabled={isSavingDesc}
            />
          </div>

          <Separator />

          {/* Comments */}
          <div className="space-y-3">
            <Label>{t("comments")}</Label>
            {comments.length === 0 && (
              <p className="text-sm text-muted-foreground">{tc("noItems")}</p>
            )}
            <div className="space-y-3 max-h-[300px] overflow-y-auto">
              {comments.map((comment) => (
                <div key={comment.id} className="rounded-lg bg-muted/50 p-3">
                  <div className="flex items-center gap-2 mb-1">
                    <Avatar className="h-5 w-5">
                      <AvatarFallback className="text-[9px]">
                        {comment.display_name.charAt(0).toUpperCase()}
                      </AvatarFallback>
                    </Avatar>
                    <span className="text-xs font-medium">{comment.display_name}</span>
                    <span className="text-xs text-muted-foreground">
                      {new Date(comment.created_at).toLocaleDateString()}
                    </span>
                  </div>
                  <p className="text-sm">{comment.content}</p>
                </div>
              ))}
            </div>
            <div className="flex gap-2">
              <Input
                value={commentText}
                onChange={(e) => setCommentText(e.target.value)}
                onKeyDown={(e) => { if (e.key === "Enter") handleAddComment(); }}
                placeholder={t("addComment")}
                className="min-h-[44px]"
              />
              <Button
                size="sm"
                onClick={handleAddComment}
                disabled={!commentText.trim()}
                className="min-h-[44px]"
              >
                {t("send")}
              </Button>
            </div>
          </div>

          <Separator />

          {/* Delete */}
          <Button
            variant="destructive"
            size="sm"
            className="w-full min-h-[44px]"
            onClick={handleDelete}
          >
            {t("deleteTask")}
          </Button>
        </div>
      </SheetContent>
    </Sheet>
  );
}
