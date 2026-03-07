"use client";

import { useTranslations } from "next-intl";
import { useAuthStore } from "@/stores/authStore";
import { LocaleSwitcher } from "./LocaleSwitcher";
import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import { Sheet, SheetContent, SheetTrigger } from "@/components/ui/sheet";
import { Sidebar } from "./Sidebar";
import { useRouter } from "@/i18n/navigation";

export function Header({ workspaceName }: { workspaceName?: string }) {
  const t = useTranslations();
  const { user, logout } = useAuthStore();
  const router = useRouter();

  const initials = user?.display_name
    ?.split(" ")
    .map((n) => n[0])
    .join("")
    .toUpperCase()
    .slice(0, 2) || "?";

  async function handleLogout() {
    await logout();
    router.replace("/login");
  }

  return (
    <header className="sticky top-0 z-40 border-b bg-background">
      <div className="flex h-14 items-center gap-2 px-4">
        {/* Mobile menu */}
        <div className="md:hidden">
          <Sheet>
            <SheetTrigger asChild>
              <Button variant="ghost" size="sm" className="min-w-[44px] min-h-[44px]" aria-label={t("nav.menu")}>
                <svg width="20" height="20" viewBox="0 0 20 20" fill="none" stroke="currentColor" strokeWidth="2" aria-hidden="true">
                  <path d="M3 5h14M3 10h14M3 15h14" />
                </svg>
              </Button>
            </SheetTrigger>
            <SheetContent side="left" className="w-64 p-0">
              <Sidebar />
            </SheetContent>
          </Sheet>
        </div>

        <span className="font-bold text-lg">{t("common.appName")}</span>
        {workspaceName && (
          <span className="text-muted-foreground text-sm hidden sm:inline">/ {workspaceName}</span>
        )}

        <div className="ml-auto flex items-center gap-1">
          <LocaleSwitcher />
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button variant="ghost" size="sm" className="min-w-[44px] min-h-[44px] rounded-full p-0">
                <Avatar className="h-8 w-8">
                  <AvatarFallback className="text-xs">{initials}</AvatarFallback>
                </Avatar>
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end">
              <DropdownMenuItem onClick={handleLogout}>
                {t("auth.logout")}
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        </div>
      </div>
    </header>
  );
}
