"use client";

import { useEffect, useState } from "react";
import { useTranslations } from "next-intl";
import { useAuthStore } from "@/stores/authStore";
import { useWorkspaceStore, type Member } from "@/stores/workspaceStore";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import { Badge } from "@/components/ui/badge";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { toast } from "sonner";
import { UserPlus, ChevronDown, Trash2 } from "lucide-react";

const roles = ["owner", "admin", "member", "viewer"] as const;

const roleBadgeColors: Record<string, string> = {
  owner: "bg-purple-100 text-purple-700",
  admin: "bg-blue-100 text-blue-700",
  member: "bg-green-100 text-green-700",
  viewer: "bg-slate-100 text-slate-700",
};

function getInitials(name: string) {
  return name
    .split(" ")
    .map((n) => n[0])
    .join("")
    .toUpperCase()
    .slice(0, 2);
}

export default function MembersPage() {
  const t = useTranslations("members");
  const tw = useTranslations("workspace");
  const tc = useTranslations("common");

  const { tenantId } = useAuthStore();
  const { members, currentUserRole, fetchMembers, updateMemberRole, removeMember } =
    useWorkspaceStore();
  const currentUserId = useAuthStore((s) => s.user?.id);

  const [isInviteOpen, setIsInviteOpen] = useState(false);
  const [inviteEmail, setInviteEmail] = useState("");
  const [removingId, setRemovingId] = useState<string | null>(null);

  const isOwner = currentUserRole === "owner";
  const isAdmin = currentUserRole === "admin" || isOwner;

  useEffect(() => {
    if (tenantId) fetchMembers(tenantId);
  }, [tenantId, fetchMembers]);

  async function handleRoleChange(userId: string, role: string) {
    if (!tenantId) return;
    try {
      await updateMemberRole(tenantId, userId, role);
    } catch {
      toast.error(tc("error"));
    }
  }

  async function handleRemove(member: Member) {
    if (!tenantId) return;
    setRemovingId(member.user_id);
    try {
      await removeMember(tenantId, member.user_id);
    } catch {
      toast.error(tc("error"));
    } finally {
      setRemovingId(null);
    }
  }

  function roleLabel(role: string) {
    const key = `role${role.charAt(0).toUpperCase() + role.slice(1)}` as
      | "roleOwner"
      | "roleAdmin"
      | "roleMember"
      | "roleViewer";
    return tw(key);
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold">{t("members")}</h1>
        <Dialog open={isInviteOpen} onOpenChange={setIsInviteOpen}>
          <DialogTrigger asChild>
            <Button className="min-h-[44px]">
              <UserPlus className="mr-2 h-4 w-4" />
              {t("inviteMember")}
            </Button>
          </DialogTrigger>
          <DialogContent>
            <DialogHeader>
              <DialogTitle>{t("inviteMember")}</DialogTitle>
              <DialogDescription>{t("inviteDescription")}</DialogDescription>
            </DialogHeader>
            <div className="space-y-4 pt-2">
              <div className="space-y-2">
                <Label>{t("inviteEmail")}</Label>
                <Input
                  type="email"
                  value={inviteEmail}
                  onChange={(e) => setInviteEmail(e.target.value)}
                  placeholder="user@example.com"
                />
              </div>
              <p className="text-xs text-muted-foreground">{t("inviteNote")}</p>
              <div className="flex justify-end">
                <Button
                  variant="outline"
                  onClick={() => setIsInviteOpen(false)}
                >
                  {tc("close")}
                </Button>
              </div>
            </div>
          </DialogContent>
        </Dialog>
      </div>

      {/* Desktop Table */}
      <div className="hidden rounded-lg border md:block">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>{t("displayName")}</TableHead>
              <TableHead>{tc("email")}</TableHead>
              <TableHead>{tc("role")}</TableHead>
              <TableHead>{t("joinedAt")}</TableHead>
              {isAdmin && <TableHead className="w-[100px]">{tc("actions")}</TableHead>}
            </TableRow>
          </TableHeader>
          <TableBody>
            {members.map((member) => (
              <TableRow key={member.user_id}>
                <TableCell>
                  <div className="flex items-center gap-3">
                    <Avatar className="h-8 w-8">
                      <AvatarFallback className="text-xs">
                        {getInitials(member.display_name)}
                      </AvatarFallback>
                    </Avatar>
                    <span className="font-medium">{member.display_name}</span>
                  </div>
                </TableCell>
                <TableCell className="text-muted-foreground">
                  {member.email}
                </TableCell>
                <TableCell>
                  {isOwner && member.user_id !== currentUserId ? (
                    <DropdownMenu>
                      <DropdownMenuTrigger asChild>
                        <Button
                          variant="ghost"
                          size="sm"
                          className="h-auto gap-1 px-2 py-1"
                        >
                          <Badge
                            variant="secondary"
                            className={roleBadgeColors[member.role] || ""}
                          >
                            {roleLabel(member.role)}
                          </Badge>
                          <ChevronDown className="h-3 w-3" />
                        </Button>
                      </DropdownMenuTrigger>
                      <DropdownMenuContent>
                        {roles.map((r) => (
                          <DropdownMenuItem
                            key={r}
                            onClick={() =>
                              handleRoleChange(member.user_id, r)
                            }
                            disabled={r === member.role}
                          >
                            {roleLabel(r)}
                          </DropdownMenuItem>
                        ))}
                      </DropdownMenuContent>
                    </DropdownMenu>
                  ) : (
                    <Badge
                      variant="secondary"
                      className={roleBadgeColors[member.role] || ""}
                    >
                      {roleLabel(member.role)}
                    </Badge>
                  )}
                </TableCell>
                <TableCell className="text-muted-foreground">
                  {new Date(member.joined_at).toLocaleDateString()}
                </TableCell>
                {isAdmin && (
                  <TableCell>
                    {member.user_id !== currentUserId &&
                      member.role !== "owner" && (
                        <Button
                          variant="ghost"
                          size="sm"
                          className="h-8 w-8 p-0 text-destructive hover:text-destructive"
                          disabled={removingId === member.user_id}
                          onClick={() => {
                            if (confirm(t("removeConfirm"))) {
                              handleRemove(member);
                            }
                          }}
                        >
                          <Trash2 className="h-4 w-4" />
                        </Button>
                      )}
                  </TableCell>
                )}
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </div>

      {/* Mobile Cards */}
      <div className="space-y-3 md:hidden">
        {members.map((member) => (
          <div
            key={member.user_id}
            className="flex items-center gap-3 rounded-lg border p-4"
          >
            <Avatar className="h-10 w-10">
              <AvatarFallback>{getInitials(member.display_name)}</AvatarFallback>
            </Avatar>
            <div className="min-w-0 flex-1">
              <p className="truncate font-medium">{member.display_name}</p>
              <p className="truncate text-sm text-muted-foreground">
                {member.email}
              </p>
              <div className="mt-1 flex items-center gap-2">
                <Badge
                  variant="secondary"
                  className={roleBadgeColors[member.role] || ""}
                >
                  {roleLabel(member.role)}
                </Badge>
                <span className="text-xs text-muted-foreground">
                  {new Date(member.joined_at).toLocaleDateString()}
                </span>
              </div>
            </div>
            {isAdmin &&
              member.user_id !== currentUserId &&
              member.role !== "owner" && (
                <Button
                  variant="ghost"
                  size="sm"
                  className="shrink-0 text-destructive hover:text-destructive min-h-[44px] min-w-[44px]"
                  disabled={removingId === member.user_id}
                  onClick={() => {
                    if (confirm(t("removeConfirm"))) {
                      handleRemove(member);
                    }
                  }}
                >
                  <Trash2 className="h-4 w-4" />
                </Button>
              )}
          </div>
        ))}
        {members.length === 0 && (
          <p className="text-center text-muted-foreground py-8">
            {t("noMembers")}
          </p>
        )}
      </div>
    </div>
  );
}
