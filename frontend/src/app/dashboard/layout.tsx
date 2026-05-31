"use client";

import { ReactNode } from "react";
import Link from "next/link";
import { usePathname } from "next/navigation";
import { useUserStore } from "@/store/useUserStore";
import { useCurrentUser } from "@/hooks/useAuth";
import ProfileDropdown from "@/components/dashboard/ProfileDropdown";
import { Link2, Megaphone, Wallet, Shield, Droplets } from "lucide-react";
import { WebSocketProvider } from "@/providers/WebSocketProvider";
import { 
  ROLE_USER, 
  ROLE_ADVERTISER, 
  ROLE_ADMIN,
  ROUTE_DASHBOARD,
  ROUTE_LINKS,
  ROUTE_CAMPAIGNS,
  ROUTE_WALLET,
  ROUTE_FAUCET,
  ROUTE_ADMIN_DASHBOARD
} from "@/lib/constants";

interface NavItem {
  href: string;
  label: string;
  icon: ReactNode;
  roles: string[];
}

const NAV_ITEMS: NavItem[] = [
  { href: ROUTE_LINKS, label: "Links", icon: <Link2 className="mr-2 h-4 w-4" />, roles: [ROLE_USER, ROLE_ADMIN] },
  { href: ROUTE_CAMPAIGNS, label: "Campaigns", icon: <Megaphone className="mr-2 h-4 w-4" />, roles: [ROLE_ADVERTISER, ROLE_ADMIN] },
  { href: ROUTE_WALLET, label: "Wallet", icon: <Wallet className="mr-2 h-4 w-4" />, roles: [ROLE_USER, ROLE_ADVERTISER, ROLE_ADMIN] },
  { href: ROUTE_FAUCET, label: "Faucet", icon: <Droplets className="mr-2 h-4 w-4" />, roles: [ROLE_USER, ROLE_ADVERTISER, ROLE_ADMIN] },
  { href: ROUTE_ADMIN_DASHBOARD, label: "Admin", icon: <Shield className="mr-2 h-4 w-4" />, roles: [ROLE_ADMIN] },
];

export default function DashboardLayout({ children }: { children: ReactNode }) {
  const pathname = usePathname();
  const user = useUserStore((state) => state.user);
  const role = user?.role || ROLE_USER;

  useCurrentUser();

  const visibleNav = NAV_ITEMS.filter((item) => item.roles.includes(role));

  return (
    <div className="min-h-screen bg-[#0A0A0A] flex flex-col font-syne grain-overlay">
      <nav className="border-b border-white/[0.06] bg-[#0A0A0A]/85 backdrop-blur-xl sticky top-0 z-10">
        <div className="mx-auto max-w-7xl px-6 md:px-12">
          <div className="flex h-16 justify-between items-center">
            <div className="flex items-center gap-8">
              <Link href={ROUTE_DASHBOARD} className="flex items-center gap-2">
                <div className="h-8 w-8 rounded-lg bg-[#6EE7B7] flex items-center justify-center">
                  <Link2 size={16} className="text-[#0A0A0A] stroke-[2.5]" />
                </div>
                <span className="text-xl font-bold text-white tracking-tight">go-short</span>
              </Link>

              <div className="hidden sm:flex sm:space-x-4">
                {visibleNav.map((item) => (
                  <Link
                    key={item.href}
                    href={item.href}
                    className={`flex items-center px-3 py-2 rounded-lg text-sm font-medium transition-colors ${
                      pathname === item.href
                        ? "bg-white/[0.06] text-white"
                        : "text-white/50 hover:text-white hover:bg-white/[0.04]"
                    }`}
                  >
                    {item.icon}
                    {item.label}
                  </Link>
                ))}
              </div>
            </div>

            <div className="flex items-center gap-3">
              {role !== ROLE_USER && (
                <span className="text-xs font-mono-dm uppercase tracking-wider px-2 py-1 rounded-md bg-white/[0.06] text-white/50">
                  {role}
                </span>
              )}
              <ProfileDropdown />
            </div>
          </div>
        </div>
      </nav>

      <main className="flex-1 max-w-7xl w-full mx-auto px-6 md:px-12 py-8">
        <WebSocketProvider>
          {children}
        </WebSocketProvider>
      </main>
    </div>
  );
}
