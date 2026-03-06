"use client";

import { useEffect, useState } from "react";
import { useTranslations } from "next-intl";
import { useAuthStore } from "@/stores/authStore";
import { useWorkspaceStore } from "@/stores/workspaceStore";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Separator } from "@/components/ui/separator";
import { toast } from "sonner";

export default function SettingsPage() {
  const t = useTranslations("settings");
  const tc = useTranslations("common");

  const { tenantId } = useAuthStore();
  const { tenant, currentUserRole, fetchTenant, fetchMembers, updateTenant } =
    useWorkspaceStore();

  const [name, setName] = useState("");
  const [isSaving, setIsSaving] = useState(false);

  const isOwner = currentUserRole === "owner";

  useEffect(() => {
    if (tenantId) {
      fetchTenant(tenantId);
      fetchMembers(tenantId);
    }
  }, [tenantId, fetchTenant, fetchMembers]);

  useEffect(() => {
    if (tenant) setName(tenant.name);
  }, [tenant]);

  async function handleUpdateName() {
    if (!tenantId || !name.trim()) return;
    setIsSaving(true);
    try {
      await updateTenant(tenantId, name.trim());
      toast.success(t("updateSuccess"));
    } catch {
      toast.error(t("updateFailed"));
    } finally {
      setIsSaving(false);
    }
  }

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-bold">{t("settings")}</h1>

      <Card>
        <CardHeader>
          <CardTitle>{t("tenantInfo")}</CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="space-y-2">
            <Label>{t("tenantName")}</Label>
            <div className="flex gap-2">
              <Input
                value={name}
                onChange={(e) => setName(e.target.value)}
                disabled={!isOwner}
                onKeyDown={(e) => {
                  if (e.key === "Enter" && isOwner) handleUpdateName();
                }}
              />
              {isOwner && (
                <Button
                  onClick={handleUpdateName}
                  disabled={isSaving || !name.trim() || name === tenant?.name}
                  className="shrink-0"
                >
                  {isSaving ? tc("loading") : t("updateName")}
                </Button>
              )}
            </div>
          </div>
          <div className="space-y-2">
            <Label>{t("tenantSlug")}</Label>
            <Input value={tenant?.slug || ""} disabled />
            <p className="text-xs text-muted-foreground">
              {t("slugReadOnly")}
            </p>
          </div>
        </CardContent>
      </Card>

      {isOwner && (
        <Card className="border-destructive/50">
          <CardHeader>
            <CardTitle className="text-destructive">
              {t("dangerZone")}
            </CardTitle>
            <CardDescription>{t("deleteDescription")}</CardDescription>
          </CardHeader>
          <CardContent>
            <Separator className="mb-4" />
            <Button
              variant="destructive"
              disabled
            >
              {t("deleteTenant")}
            </Button>
          </CardContent>
        </Card>
      )}
    </div>
  );
}
