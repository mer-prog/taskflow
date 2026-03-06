"use client";

import { useEffect, useState } from "react";
import { useTranslations } from "next-intl";
import { useParams } from "next/navigation";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { Label } from "@/components/ui/label";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { useWorkspaceStore } from "@/stores/workspaceStore";
import { Link } from "@/i18n/navigation";
import { toast } from "sonner";
import { Plus, FolderOpen } from "lucide-react";

export default function ProjectsPage() {
  const t = useTranslations("project");
  const tc = useTranslations("common");
  const params = useParams();
  const slug = params?.slug as string;

  const { projects, fetchProjects, createProject } = useWorkspaceStore();
  const [isDialogOpen, setIsDialogOpen] = useState(false);
  const [name, setName] = useState("");
  const [description, setDescription] = useState("");
  const [isCreating, setIsCreating] = useState(false);

  useEffect(() => {
    fetchProjects();
  }, [fetchProjects]);

  async function handleCreate() {
    if (!name.trim()) return;
    setIsCreating(true);
    try {
      await createProject(name.trim(), description.trim());
      setName("");
      setDescription("");
      setIsDialogOpen(false);
    } catch {
      toast.error(tc("error"));
    } finally {
      setIsCreating(false);
    }
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold">{t("projects")}</h1>
        <Dialog open={isDialogOpen} onOpenChange={setIsDialogOpen}>
          <DialogTrigger asChild>
            <Button className="min-h-[44px]">
              <Plus className="mr-2 h-4 w-4" />
              {t("newProject")}
            </Button>
          </DialogTrigger>
          <DialogContent>
            <DialogHeader>
              <DialogTitle>{t("createProject")}</DialogTitle>
            </DialogHeader>
            <div className="space-y-4 pt-2">
              <div className="space-y-2">
                <Label>{t("projectName")}</Label>
                <Input
                  value={name}
                  onChange={(e) => setName(e.target.value)}
                  placeholder={t("projectName")}
                  onKeyDown={(e) => {
                    if (e.key === "Enter") handleCreate();
                  }}
                />
              </div>
              <div className="space-y-2">
                <Label>{t("projectDescription")}</Label>
                <Textarea
                  value={description}
                  onChange={(e) => setDescription(e.target.value)}
                  placeholder={t("projectDescription")}
                  rows={3}
                />
              </div>
              <div className="flex justify-end gap-2">
                <Button
                  variant="outline"
                  onClick={() => setIsDialogOpen(false)}
                >
                  {tc("cancel")}
                </Button>
                <Button
                  onClick={handleCreate}
                  disabled={!name.trim() || isCreating}
                >
                  {isCreating ? tc("loading") : tc("create")}
                </Button>
              </div>
            </div>
          </DialogContent>
        </Dialog>
      </div>

      {projects.length === 0 ? (
        <Card>
          <CardContent className="flex flex-col items-center justify-center py-12">
            <FolderOpen className="mb-4 h-12 w-12 text-muted-foreground" />
            <p className="text-lg font-medium">{t("noProjects")}</p>
            <p className="text-sm text-muted-foreground">
              {t("createFirst")}
            </p>
          </CardContent>
        </Card>
      ) : (
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {projects.map((project) => (
            <Link
              key={project.id}
              href={`/ws/${slug}/p/${project.id}/board`}
            >
              <Card className="h-full transition-colors hover:bg-accent/50 cursor-pointer">
                <CardHeader>
                  <CardTitle className="flex items-center gap-2 text-base">
                    <FolderOpen className="h-5 w-5 shrink-0 text-muted-foreground" />
                    {project.name}
                  </CardTitle>
                  {project.description && (
                    <CardDescription className="line-clamp-2">
                      {project.description}
                    </CardDescription>
                  )}
                </CardHeader>
                <CardContent>
                  <p className="text-xs text-muted-foreground">
                    {new Date(project.created_at).toLocaleDateString()}
                  </p>
                </CardContent>
              </Card>
            </Link>
          ))}
        </div>
      )}
    </div>
  );
}
