"use client";

import { useEffect } from "react";
import { useTranslations } from "next-intl";
import { useParams } from "next/navigation";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { useDashboardStore } from "@/stores/dashboardStore";
import { useWorkspaceStore } from "@/stores/workspaceStore";
import { Link } from "@/i18n/navigation";
import {
  CheckCircle2,
  Clock,
  AlertTriangle,
  ListTodo,
  ArrowRight,
  FolderOpen,
} from "lucide-react";

const priorityColors: Record<string, string> = {
  low: "bg-slate-100 text-slate-700",
  medium: "bg-blue-100 text-blue-700",
  high: "bg-orange-100 text-orange-700",
  critical: "bg-red-100 text-red-700",
};

export default function DashboardPage() {
  const t = useTranslations("dashboard");
  const tb = useTranslations("board");
  const tc = useTranslations("common");
  const params = useParams();
  const slug = params?.slug as string;

  const { summary, myTasks, overdueTasks, isLoading, fetchAll } =
    useDashboardStore();
  const { projects, fetchProjects } = useWorkspaceStore();

  useEffect(() => {
    fetchAll();
    fetchProjects();
  }, [fetchAll, fetchProjects]);

  if (isLoading) {
    return (
      <div className="flex items-center justify-center py-20">
        <p className="text-muted-foreground">{tc("loading")}</p>
      </div>
    );
  }

  const totalTasks = summary?.total_tasks ?? 0;
  const overdueCnt = summary?.overdue_tasks ?? 0;
  const byColumn = summary?.by_column ?? [];
  const completedTasks = byColumn
    .filter((c) => {
      const name = c.column_name.toLowerCase();
      return name.includes("done") || name.includes("complete") || name.includes("完了");
    })
    .reduce((sum, c) => sum + c.task_count, 0);
  const inProgressTasks = totalTasks - completedTasks - overdueCnt;

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-bold">{t("dashboard")}</h1>

      {/* Summary Cards */}
      <div className="grid grid-cols-2 gap-4 sm:grid-cols-4">
        <Card>
          <CardContent className="pt-6">
            <div className="flex items-center gap-3">
              <div className="rounded-lg bg-blue-100 p-2">
                <ListTodo className="h-5 w-5 text-blue-600" />
              </div>
              <div>
                <p className="text-sm text-muted-foreground">
                  {t("totalTasks")}
                </p>
                <p className="text-2xl font-bold">{totalTasks}</p>
              </div>
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="pt-6">
            <div className="flex items-center gap-3">
              <div className="rounded-lg bg-green-100 p-2">
                <CheckCircle2 className="h-5 w-5 text-green-600" />
              </div>
              <div>
                <p className="text-sm text-muted-foreground">
                  {t("completedTasks")}
                </p>
                <p className="text-2xl font-bold">{completedTasks}</p>
              </div>
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="pt-6">
            <div className="flex items-center gap-3">
              <div className="rounded-lg bg-yellow-100 p-2">
                <Clock className="h-5 w-5 text-yellow-600" />
              </div>
              <div>
                <p className="text-sm text-muted-foreground">
                  {t("inProgress")}
                </p>
                <p className="text-2xl font-bold">
                  {inProgressTasks > 0 ? inProgressTasks : 0}
                </p>
              </div>
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="pt-6">
            <div className="flex items-center gap-3">
              <div className="rounded-lg bg-red-100 p-2">
                <AlertTriangle className="h-5 w-5 text-red-600" />
              </div>
              <div>
                <p className="text-sm text-muted-foreground">
                  {t("overdueTasks")}
                </p>
                <p className="text-2xl font-bold">{overdueCnt}</p>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>

      <div className="grid gap-6 md:grid-cols-2">
        {/* My Tasks */}
        <Card>
          <CardHeader>
            <CardTitle className="text-lg">{t("myTasks")}</CardTitle>
          </CardHeader>
          <CardContent>
            {myTasks.length === 0 ? (
              <p className="text-sm text-muted-foreground">
                {t("noMyTasks")}
              </p>
            ) : (
              <div className="space-y-3">
                {myTasks.map((task) => (
                  <div
                    key={task.id}
                    className="flex items-center justify-between rounded-lg border p-3"
                  >
                    <div className="min-w-0 flex-1">
                      <p className="truncate text-sm font-medium">
                        {task.title}
                      </p>
                      <p className="text-xs text-muted-foreground">
                        {task.column_name}
                        {task.board_name && ` · ${task.board_name}`}
                      </p>
                    </div>
                    <div className="flex shrink-0 items-center gap-2">
                      {task.priority && (
                        <Badge
                          variant="secondary"
                          className={
                            priorityColors[task.priority] || ""
                          }
                        >
                          {tb(
                            `priority${task.priority.charAt(0).toUpperCase() + task.priority.slice(1)}` as
                              | "priorityLow"
                              | "priorityMedium"
                              | "priorityHigh"
                              | "priorityCritical"
                          )}
                        </Badge>
                      )}
                      {task.due_date && (
                        <Badge
                          variant="outline"
                          className={
                            new Date(task.due_date) < new Date()
                              ? "border-red-300 text-red-600"
                              : ""
                          }
                        >
                          {new Date(task.due_date).toLocaleDateString()}
                        </Badge>
                      )}
                    </div>
                  </div>
                ))}
              </div>
            )}
          </CardContent>
        </Card>

        {/* Overdue Tasks */}
        <Card>
          <CardHeader>
            <CardTitle className="text-lg text-red-600">
              {t("overdueTasks")}
            </CardTitle>
          </CardHeader>
          <CardContent>
            {overdueTasks.length === 0 ? (
              <p className="text-sm text-muted-foreground">
                {t("noOverdueTasks")}
              </p>
            ) : (
              <div className="space-y-3">
                {overdueTasks.map((task) => (
                  <div
                    key={task.id}
                    className="flex items-center justify-between rounded-lg border border-red-200 bg-red-50 p-3"
                  >
                    <div className="min-w-0 flex-1">
                      <p className="truncate text-sm font-medium text-red-900">
                        {task.title}
                      </p>
                      <p className="text-xs text-red-600">
                        {task.column_name}
                        {task.board_name && ` · ${task.board_name}`}
                      </p>
                    </div>
                    {task.due_date && (
                      <Badge
                        variant="destructive"
                        className="shrink-0"
                      >
                        {new Date(task.due_date).toLocaleDateString()}
                      </Badge>
                    )}
                  </div>
                ))}
              </div>
            )}
          </CardContent>
        </Card>
      </div>

      {/* Quick Links - Projects */}
      <Card>
        <CardHeader className="flex flex-row items-center justify-between">
          <CardTitle className="text-lg">{t("quickLinks")}</CardTitle>
          <Link
            href={`/ws/${slug}/projects`}
            className="flex items-center gap-1 text-sm text-muted-foreground hover:text-foreground"
          >
            {tc("viewAll")}
            <ArrowRight className="h-4 w-4" />
          </Link>
        </CardHeader>
        <CardContent>
          {projects.length === 0 ? (
            <p className="text-sm text-muted-foreground">
              {tc("noItems")}
            </p>
          ) : (
            <div className="grid gap-3 sm:grid-cols-2 md:grid-cols-3">
              {projects.slice(0, 6).map((project) => (
                <Link
                  key={project.id}
                  href={`/ws/${slug}/projects`}
                  className="flex items-center gap-3 rounded-lg border p-3 transition-colors hover:bg-accent"
                >
                  <FolderOpen className="h-5 w-5 shrink-0 text-muted-foreground" />
                  <div className="min-w-0">
                    <p className="truncate text-sm font-medium">
                      {project.name}
                    </p>
                    {project.description && (
                      <p className="truncate text-xs text-muted-foreground">
                        {project.description}
                      </p>
                    )}
                  </div>
                </Link>
              ))}
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
