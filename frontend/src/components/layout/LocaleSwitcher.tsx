"use client";

import { useLocale } from "next-intl";
import { usePathname, useRouter } from "@/i18n/navigation";
import { Button } from "@/components/ui/button";

export function LocaleSwitcher() {
  const locale = useLocale();
  const router = useRouter();
  const pathname = usePathname();

  function switchLocale() {
    const next = locale === "ja" ? "en" : "ja";
    router.replace(pathname, { locale: next });
  }

  return (
    <Button variant="ghost" size="sm" onClick={switchLocale} className="min-w-[44px] min-h-[44px]">
      {locale === "ja" ? "EN" : "JA"}
    </Button>
  );
}
