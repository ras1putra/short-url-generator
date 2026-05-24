"use client";

import RequireRole from "@/components/auth/RequireRole";
import CreateLinkForm from "@/components/links/CreateLinkForm";
import LinkTable from "@/components/links/LinkTable";
import DashboardGlobe from "@/components/links/DashboardGlobe";
import { Link2 } from "lucide-react";
import { ROLE_USER, ROLE_ADMIN } from "@/lib/constants";

export default function LinksPage() {
  return (
    <RequireRole roles={[ROLE_USER, ROLE_ADMIN]}>
      <div className="mb-8">
        <div className="flex items-center gap-3 mb-2">
          <div className="h-8 w-8 rounded-lg bg-[#6EE7B7]/10 flex items-center justify-center">
            <Link2 size={16} className="text-[#6EE7B7]" />
          </div>
          <h1 className="text-3xl font-black tracking-tight text-white">Links</h1>
        </div>
        <p className="mt-2 text-white/50 font-mono-dm text-sm">{"// Create, manage, and track your shortened URLs"}</p>
      </div>

      <DashboardGlobe />

      <div className="mt-8">
        <CreateLinkForm />
      </div>

      <div className="mt-12">
        <h2 className="text-xl font-bold text-white/90 mb-6">Your Links</h2>
        <LinkTable />
      </div>
    </RequireRole>
  );
}
