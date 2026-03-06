"use client";

import { Header } from "@/components/layout/Header";
import { Sidebar } from "@/components/layout/Sidebar";
import { useParams } from "next/navigation";

export default function WorkspaceLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  const params = useParams();
  const slug = params?.slug as string;

  return (
    <div className="flex min-h-screen flex-col">
      <Header workspaceName={slug} />
      <div className="flex flex-1">
        <aside className="hidden w-56 shrink-0 border-r md:block">
          <Sidebar />
        </aside>
        <main className="flex-1 p-4 sm:p-6">{children}</main>
      </div>
    </div>
  );
}
