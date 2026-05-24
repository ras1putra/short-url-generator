"use client";

import { useEffect } from "react";
import { useRouter } from "next/navigation";
import { useUserStore } from "@/store/useUserStore";
import { Loader2 } from "lucide-react";
import { ROLE_ADVERTISER, ROLE_ADMIN, ROUTE_CAMPAIGNS, ROUTE_ADMIN_DASHBOARD, ROUTE_LINKS } from "@/lib/constants";

export default function DashboardPage() {
  const router = useRouter();
  const user = useUserStore((state) => state.user);

  useEffect(() => {
    if (!user) return;

    if (user.role === ROLE_ADVERTISER) {
      router.replace(ROUTE_CAMPAIGNS);
    } else if (user.role === ROLE_ADMIN) {
      router.replace(ROUTE_ADMIN_DASHBOARD);
    } else {
      router.replace(ROUTE_LINKS);
    }
  }, [user, router]);


  return (
    <div className="flex items-center justify-center py-32">
      <Loader2 className="animate-spin h-8 w-8 text-white/30" />
    </div>
  );
}
