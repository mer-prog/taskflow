"use client";

import { useTranslations } from "next-intl";
import { useParams } from "next/navigation";
import { Link } from "@/i18n/navigation";
import { usePathname } from "@/i18n/navigation";
import { cn } from "@/lib/utils";

const navItems = [
  { key: "dashboard", href: "" },
  { key: "projects", href: "/projects" },
  { key: "members", href: "/members" },
  { key: "settings", href: "/settings" },
] as const;

export function Sidebar() {
  const t = useTranslations("nav");
  const params = useParams();
  const pathname = usePathname();
  const slug = params?.slug as string;
  const basePath = `/ws/${slug}`;

  return (
    <nav className="flex flex-col gap-1 p-4">
      <span className="mb-4 px-2 text-xs font-semibold uppercase text-muted-foreground tracking-wider">
        {t("dashboard")}
      </span>
      {navItems.map((item) => {
        const href = `${basePath}${item.href}`;
        const isActive = pathname === href;
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
            {t(item.key)}
          </Link>
        );
      })}
    </nav>
  );
}
