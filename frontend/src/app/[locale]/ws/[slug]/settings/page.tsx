"use client";

import { useTranslations } from "next-intl";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";

export default function SettingsPage() {
  const t = useTranslations("workspace");
  const tc = useTranslations("common");

  return (
    <div>
      <h1 className="text-2xl font-bold mb-6">{t("settings")}</h1>
      <Card>
        <CardHeader>
          <CardTitle>{t("settings")}</CardTitle>
        </CardHeader>
        <CardContent>
          <p className="text-muted-foreground">{tc("comingSoon")}</p>
        </CardContent>
      </Card>
    </div>
  );
}
