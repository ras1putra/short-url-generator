"use client";

import { ReactNode } from "react";
import Link from "next/link";
import { usePathname } from "next/navigation";
import { useLogout } from "@/hooks/useAuth";
import { LayoutDashboard, LogOut, Link2 } from "lucide-react";

export default function DashboardLayout({ children }: { children: ReactNode }) {
  const pathname = usePathname();
  const logoutMutation = useLogout();

  return (
    <div className="min-h-screen bg-[#0A0A0A] flex flex-col font-syne grain-overlay">
      <nav className="border-b border-white/[0.06] bg-[#0A0A0A]/85 backdrop-blur-xl sticky top-0 z-10">
        <div className="mx-auto max-w-7xl px-6 md:px-12">
          <div className="flex h-16 justify-between items-center">
            <div className="flex items-center gap-8">
              <Link href="/dashboard" className="flex items-center gap-2">
                <div className="h-8 w-8 rounded-lg bg-[#6EE7B7] flex items-center justify-center">
                  <Link2 size={16} className="text-[#0A0A0A] stroke-[2.5]" />
                </div>
                <span className="text-xl font-bold text-white tracking-tight">go-short</span>
              </Link>

              <div className="hidden sm:flex sm:space-x-4">
                <Link
                  href="/dashboard"
                  className={`flex items-center px-3 py-2 rounded-lg text-sm font-medium transition-colors ${
                    pathname === "/dashboard"
                      ? "bg-white/[0.06] text-white"
                      : "text-white/50 hover:text-white hover:bg-white/[0.04]"
                  }`}
                >
                  <LayoutDashboard className="mr-2 h-4 w-4" />
                  Dashboard
                </Link>
              </div>
            </div>

            <div className="flex items-center">
              <button
                onClick={() => logoutMutation.mutate()}
                disabled={logoutMutation.isPending}
                className="flex items-center text-white/50 hover:text-white px-3 py-2 rounded-lg text-sm font-medium transition-colors cursor-pointer"
              >
                <LogOut className="mr-2 h-4 w-4" />
                Sign out
              </button>
            </div>
          </div>
        </div>
      </nav>

      <main className="flex-1 max-w-7xl w-full mx-auto px-6 md:px-12 py-8">
        {children}
      </main>
    </div>
  );
}