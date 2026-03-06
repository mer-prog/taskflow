"use client";

import { useEffect } from "react";
import { useParams } from "next/navigation";
import { Header } from "@/components/layout/Header";
import { Sidebar } from "@/components/layout/Sidebar";
import { useAuthStore } from "@/stores/authStore";
import { useWorkspaceStore } from "@/stores/workspaceStore";

export default function WorkspaceLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  const params = useParams();
  const slug = params?.slug as string;

  const { setTenantId } = useAuthStore();
  const { tenant, fetchTenants, fetchTenant } = useWorkspaceStore();

  useEffect(() => {
    async function resolveTenant() {
      try {
        const tenants = await fetchTenants();
        const matched = tenants.find((t) => t.slug === slug);
        if (matched) {
          setTenantId(matched.id);
          fetchTenant(matched.id);
        }
      } catch {
        // ignore
      }
    }
    if (slug) resolveTenant();
  }, [slug, setTenantId, fetchTenants, fetchTenant]);

  return (
    <div className="flex min-h-screen flex-col">
      <Header workspaceName={tenant?.name || slug} />
      <div className="flex flex-1">
        <aside className="hidden w-56 shrink-0 border-r md:block">
          <Sidebar />
        </aside>
        <main className="flex-1 p-4 sm:p-6">{children}</main>
      </div>
    </div>
  );
}
