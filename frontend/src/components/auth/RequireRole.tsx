"use client";

import { useUserStore } from "@/store/useUserStore";
import { useRouter } from "next/navigation";
import { useEffect, ReactNode } from "react";
import { ROLE_USER } from "@/lib/constants";

interface RequireRoleProps {
  roles: string[];
  children: ReactNode;
  fallback?: ReactNode;
}

export default function RequireRole({ roles, children, fallback }: RequireRoleProps) {
  const user = useUserStore((state) => state.user);
  const router = useRouter();

  useEffect(() => {
    if (!user) return;
    if (!roles.includes(user.role || ROLE_USER)) {
      router.replace("/dashboard");
    }
  }, [user, roles, router]);

  if (!user) return null;

  if (!roles.includes(user.role || ROLE_USER)) {
    return fallback ?? null;
  }


  return <>{children}</>;
}
