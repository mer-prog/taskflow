"use client";

import { useEffect } from "react";
import { useTranslations } from "next-intl";
import { useParams } from "next/navigation";
import { Link } from "@/i18n/navigation";
import { usePathname } from "@/i18n/navigation";
import { cn } from "@/lib/utils";
import { useWorkspaceStore } from "@/stores/workspaceStore";
import {
  LayoutDashboard,
  FolderOpen,
  Users,
  Settings,
  ArrowRight,
} from "lucide-react";

const navItems = [
  { key: "dashboard", href: "", icon: LayoutDashboard },
  { key: "projects", href: "/projects", icon: FolderOpen },
  { key: "members", href: "/members", icon: Users },
  { key: "settings", href: "/settings", icon: Settings },
] as const;

export function Sidebar() {
  const t = useTranslations("nav");
  const tc = useTranslations("common");
  const params = useParams();
  const pathname = usePathname();
  const slug = params?.slug as string;
  const basePath = `/ws/${slug}`;

  const { projects, fetchProjects } = useWorkspaceStore();

  useEffect(() => {
    fetchProjects();
  }, [fetchProjects]);

  return (
    <nav className="flex flex-col gap-1 p-4">
      <span className="mb-2 px-2 text-xs font-semibold uppercase text-muted-foreground tracking-wider">
        {t("dashboard")}
      </span>
      {navItems.map((item) => {
        const href = `${basePath}${item.href}`;
        const isActive =
          item.href === ""
            ? pathname === href
            : pathname.startsWith(href);
        const Icon = item.icon;
        return (
          <Link
            key={item.key}
            href={href}
            className={cn(
              "flex items-center gap-3 rounded-md px-3 py-2 text-sm font-medium min-h-[44px] transition-colors",
              isActive
                ? "bg-accent text-accent-foreground"
                : "text-muted-foreground hover:bg-accent hover:text-accent-foreground"
            )}
          >
            <Icon className="h-4 w-4 shrink-0" />
            {t(item.key)}
          </Link>
        );
      })}

      {/* Project List */}
      {projects.length > 0 && (
        <>
          <div className="mt-6 mb-2 flex items-center justify-between px-2">
            <span className="text-xs font-semibold uppercase text-muted-foreground tracking-wider">
              {t("projects")}
            </span>
            {projects.length > 5 && (
              <Link
                href={`${basePath}/projects`}
                className="flex items-center gap-0.5 text-xs text-muted-foreground hover:text-foreground"
              >
                {tc("viewAll")}
                <ArrowRight className="h-3 w-3" />
              </Link>
            )}
          </div>
          {projects.slice(0, 5).map((project) => {
            const href = `${basePath}/p/${project.id}/board`;
            const isActive = pathname.startsWith(
              `${basePath}/p/${project.id}`
            );
            return (
              <Link
                key={project.id}
                href={href}
                className={cn(
                  "flex items-center gap-3 rounded-md px-3 py-2 text-sm min-h-[44px] transition-colors",
                  isActive
                    ? "bg-accent text-accent-foreground font-medium"
                    : "text-muted-foreground hover:bg-accent hover:text-accent-foreground"
                )}
              >
                <FolderOpen className="h-4 w-4 shrink-0" />
                <span className="truncate">{project.name}</span>
              </Link>
            );
          })}
        </>
      )}
    </nav>
  );
}
