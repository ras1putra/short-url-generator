"use client";

import RequireRole from "@/components/auth/RequireRole";
import { Shield } from "lucide-react";
import { ROLE_ADMIN } from "@/lib/constants";

export default function AdminPage() {
  return (
    <RequireRole roles={[ROLE_ADMIN]}>

      <div className="mb-8">
        <div className="flex items-center gap-3 mb-2">
          <div className="h-8 w-8 rounded-lg bg-red-500/10 flex items-center justify-center">
            <Shield size={16} className="text-red-400" />
          </div>
          <h1 className="text-3xl font-black tracking-tight text-white">Admin</h1>
        </div>
        <p className="mt-2 text-white/50 font-mono-dm text-sm">{"// Platform settings and management"}</p>
      </div>

      <div className="rounded-2xl border border-white/[0.08] bg-white/[0.02] p-12 text-center">
        <Shield size={40} className="mx-auto mb-4 text-white/20" />
        <h2 className="text-xl font-bold text-white/60 mb-2">Admin Panel Coming Soon</h2>
        <p className="text-white/40 text-sm max-w-md mx-auto">
          You&apos;ll be able to manage platform commission rates, users, and system-wide settings here.
        </p>
      </div>
    </RequireRole>
  );
}
